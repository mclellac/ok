package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	//ui "github.com/mclellac/amity/lib/ui"
	pb "github.com/mclellac/ok/protos/post"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

var conf Config

type Config struct {
	Domain        string `yaml:"domain"`
	Port          string `yaml:"port"`
	ServerLogfile string `yaml:"server_logfile"`
	PubKey        string `yaml:"pub_key"`
	CertPEM       string `yaml:"cert_pem"`
	CertCSR       string `yaml:"cert_csr"`
	CertCRT       string `yaml:"cert_crt"`
	TypeDB        string `yaml:"type_db"`
	DBConn        string `yaml:"db_connect"`
	DBUsername    string `yaml:"db_username"`
	DBPassword    string `yaml:"db_password"`
	DBHostname    string `yaml:"db_hostname"`
	DBName        string `yaml:"db_name"`
	DBLogfile     string `yaml:"db_logfile"`
}

type Post struct {
	ID      int64  `gorm:"primary_key"`
	Created int32  `gorm:"size:25"`
	Title   string `gorm:"type:varchar(100)"`
	Article string `gorm:"type:varchar(5000)"`
}

type postService struct {
	post []*pb.Post
	m    sync.Mutex
	DB   *gorm.DB
}

func (ps *postService) Delete(c context.Context, req *pb.Post) (*pb.Response, error) {
	ps.m.Lock()
	defer ps.m.Unlock()

	if ps.DB.First(req).RecordNotFound() {
		fmt.Println("unable to find the requested post.")
		return &pb.Response{
			Error: fmt.Sprintf("you sure there is a post with the ID %d, sport?", int64(req.Id)),
		}, nil
	} else {
		ps.DB.Delete(req)
	}

	return &pb.Response{
		Message: fmt.Sprintf("post with ID %d has been blasted into oblivion", int64(req.Id)),
	}, nil
}

func (ps *postService) Add(c context.Context, req *pb.Post) (*pb.Response, error) {
	ps.m.Lock()
	defer ps.m.Unlock()
	ps.post = append(ps.post, req)

	req.Created = int32(time.Now().Unix())
	ps.DB.Save(&req)

	return new(pb.Response), nil
}

func (ps *postService) List(req *pb.Request, stream pb.Service_ListServer) error {
	var post []*pb.Post

	ps.m.Lock()
	defer ps.m.Unlock()

	ps.DB.Order("created desc").Find(&post)

	for _, r := range post {
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
		conf.DBName + "?charset=utf8&parseTime=True&loc=Local"

	db, err := gorm.Open(conf.TypeDB, connectionString)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}
	defer db.Close()

	dblog, err := os.OpenFile(conf.DBLogfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer dblog.Close()

	// Disable table name's pluralization
	db.SingularTable(true)
	// Enable Logger
	db.LogMode(true)
	db.SetLogger(log.New(dblog, "\r\n", 0))

	if !db.HasTable("post") {
		//db.AutoMigrate(&Post{})
		db.CreateTable(&Post{})
	}

	lis, err := net.Listen("tcp", conf.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterServiceServer(s, &postService{DB: db})
	fmt.Println("Server started on port", conf.Port)
	s.Serve(lis)
}
