package main

import (
    "io"

    ui "github.com/mclellac/amity/lib/ui"
    pb "github.com/mclellac/ok/protos/post"

    "github.com/mattn/sc"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/grpclog"
)

const (
    address    = "localhost:50051"
    defaultNum = 0
)

//const address = "127.0.0.1:11111"

func add(client pb.PostServiceClient, title string, article string) error {
    post := &pb.Post{
        Title:   title,
        Article: article,
    }
    _, err := client.AddPost(context.Background(), post)
    return err
}

func list(client pb.PostServiceClient) error {
    stream, err := client.ListPost(context.Background(), new(pb.RequestType))
    if err != nil {
        return err
    }
    for {
        post, err := stream.Recv()
        if err == io.EOF {
            break
        }
        if err != nil {
            return err
        }
        grpclog.Printf("%+v %+v %+v", ui.Cyan, post, ui.Reset)
    }
    return nil
}

func main() {
    conn, err := grpc.Dial(address, grpc.WithInsecure())
    if err != nil {
        grpclog.Fatalf("failed to dial: %v", err)
    }
    defer conn.Close()
    client := pb.NewPostServiceClient(conn)

    (&sc.Cmds{
        {
            Name: "ls",
            Desc: "ls: list posts",
            Run: func(c *sc.C, args []string) error {
                return list(client)
            },
        },
        {
            Name: "add",
            Desc: "add \"title\" \"article\": add post",
            Run: func(c *sc.C, args []string) error {
                if len(args) != 2 {
                    return sc.UsageError
                }
                title := args[0]
                article := args[1]
                if err != nil {
                    return err
                }
                return add(client, title, article)
            },
        },
    }).Run(&sc.C{
        Desc: "The Overkill client",
    })
}
