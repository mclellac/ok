package main

import (
    "fmt"
    "log"
    "net"
    "os"
    "sync"
    "time"

    pb "github.com/mclellac/ok/protos/post"

    _ "github.com/go-sql-driver/mysql"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "gopkg.in/yaml.v2"
)

var conf Config

type Config struct {
    Domain     string `yaml:"domain"`
    Port       string `yaml:"port"`
    PubKey     string `yaml:"pub_key"`
    CertPEM    string `yaml:"cert_pem"`
    CertCSR    string `yaml:"cert_csr"`
    CertCRT    string `yaml:"cert_crt"`
    TypeDB     string `yaml:"type_db"`
    DBConn     string `yaml:"db_connect"`
    DBUsername string `yaml:"db_username"`
    DBPassword string `yaml:"db_password"`
    DBHostname string `yaml:"db_hostname"`
    DBName     string `yaml:"db_name"`
}

type Post struct {
    ID      int64     `gorm:"primary_key"`
    Created time.Time `gorm:"size:25"`
    Title   string    `gorm:"type:varchar(100)"`
    Article string    `gorm:"type:varchar(5000)"`
}

type postService struct {
    post []*pb.Post
    m    sync.Mutex
    DB   *gorm.DB
}

func (ps *postService) AddPost(c context.Context, p *pb.Post) (*pb.ResponseType, error) {
    ps.m.Lock()
    defer ps.m.Unlock()
    ps.post = append(ps.post, p)

    p.Created = int32(time.Now().Unix())

    fmt.Printf("ps.post = %+v\n", ps.post)
    fmt.Printf("p = %+v\n", p)
    fmt.Printf("&Post{} = %+v\n", &Post{})

    ps.DB.Save(&p)
    return new(pb.ResponseType), nil
}

func (ps *postService) ListPost(p *pb.RequestType, stream pb.PostService_ListPostServer) error {
    ps.m.Lock()
    defer ps.m.Unlock()
    for _, r := range ps.post {
        if err := stream.Send(r); err != nil {
            return err
        }
    }
    return nil
}

func (c *Config) Init() {
    file, err := os.Open("postd.yaml")
    if err != nil {
        log.Fatalf("failed to open postd.yaml server config file: %v", err)
    }
    defer file.Close()

    stat, _ := file.Stat()

    bs := make([]byte, stat.Size())
    _, err = file.Read(bs)
    if err != nil {
        log.Fatalf("failed to open settings file: %v", err)
    }

    err = yaml.Unmarshal(bs, &c)
    if err != nil {
        log.Fatalf("error: %v", err)
    }
}

func main() {
    conf.Init()

    connectionString := conf.DBUsername + ":" +
        conf.DBPassword + "@tcp(" +
        conf.DBHostname + ":3306)/" +
        conf.DBName + "?charset=utf8&parseTime=True"

    db, err := gorm.Open(conf.TypeDB, connectionString)
    if err != nil {
        log.Fatalf("failed to open DB: %v", err)
    }
    defer db.Close()

    // Disable table name's pluralization
    db.SingularTable(true)
    // Enable Logger
    db.LogMode(true)
    db.SetLogger(log.New(os.Stdout, "\r\n", 0))

    db.AutoMigrate(&Post{})

    if !db.HasTable("posts") {
        db.CreateTable(&Post{})
    }

    lis, err := net.Listen("tcp", conf.Port)
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    s := grpc.NewServer()

    pb.RegisterPostServiceServer(s, &postService{DB: db})
    fmt.Println("Server started on port", conf.Port)
    s.Serve(lis)
}
