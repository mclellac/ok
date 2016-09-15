package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sync"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"

	pb "github.com/mclellac/ok/protos/post"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

var conf Config

type Config struct {
	Domain  string `yaml:"domain"`
	Port    string `yaml:"port"`
	PubKey  string `yaml:"pub_key"`
	CertPEM string `yaml:"cert_pem"`
	CertCSR string `yaml:"cert_csr"`
	CertCRT string `yaml:"cert_crt"`
	TypeDB  string `yaml:"type_db"`
	DBConn  string `yaml:"db_connect"`
}

type Post struct {
	ID      int64  `gorm:"primary_key"`
	Title   string `gorm:"type:varchar(100)"`
	Article string `gorm:"type:varchar(5000)"`
}

type server struct {
	DB *gorm.DB
}

type postService struct {
	post []*pb.Post
	m    sync.Mutex
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

func (ps *postService) AddPost(c context.Context, p *pb.Post) (*pb.ResponseType, error) {
	ps.m.Lock()
	defer ps.m.Unlock()
	ps.post = append(ps.post, p)
	return new(pb.ResponseType), nil
}

func (c *Config) Init() {
	file, err := os.Open("settings.yaml")
	if err != nil {
		log.Fatalf("failed to open settings file: %v", err)
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

	db, err := gorm.Open(conf.TypeDB, conf.DBConn)

	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}

	defer db.Close()

	if !db.HasTable("posts") {
		db.CreateTable(&Post{})
	}

	lis, err := net.Listen("tcp", conf.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterPostServiceServer(s, new(postService))

	//pb.RegisterSenderServer(s, &server{DB: db})

	fmt.Println("Server started on port", conf.Port)
	s.Serve(lis)
}
