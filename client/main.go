package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strconv"

	ui "github.com/mclellac/amity/lib/ui"
	pb "github.com/mclellac/ok/protos/post"

	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	address    = "localhost:50051"
	defaultNum = 0
)

func add(client pb.ServiceClient, title string, article string) error {
	post := &pb.Post{
		Title:   title,
		Article: article,
	}
	res, err := client.Add(context.Background(), post)
	fmt.Println(res)

	return err
}

func delete(client pb.ServiceClient, id int64) error {
	post := &pb.Post{
		Id: id,
	}

	res, err := client.Delete(context.Background(), post)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res)

	return nil
}

func list(client pb.ServiceClient) error {
	stream, err := client.List(context.Background(), new(pb.Request))
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
	client := pb.NewServiceClient(conn)

	app := cli.NewApp()
	app.Name = "ok"
	app.Usage = "The Overkill client."
	app.Version = "0.0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "host", Value: "localhost:50051", Usage: "okd server host"},
	}

	app.Commands = []cli.Command{
		{
			Name:        "add",
			Usage:       "Create a new post.",
			Description: "Adds new article to the database.\n\nEXAMPLE:\n   $ ok add \"Test Title\" \"Test article body...\"",
			ArgsUsage:   "[\"post title\"] [\"post body\"]",
			Action: func(c *cli.Context) error {
				if len(c.Args()) != 2 {
					fmt.Println("y'might want to double check your command there, cowgirl.")
				}

				title := c.Args().Get(0)
				article := c.Args().Get(1)
				if err != nil {
					return err
				}
				return add(client, title, article)
			},
		},
		{
			Name:        "ls",
			Usage:       "List all posts.",
			Description: "Displays the IDs and titles of posts on the server.\n\nEXAMPLE:\n   $ ok ls",
			Action: func(c *cli.Context) error {
				return list(client)
			},
		},
		{
			Name:        "rm",
			Usage:       "Delete a post.",
			Description: "Remove the post with the supplied ID from the server.\n\nEXAMPLE:\n   $ ok rm 2",
			ArgsUsage:   "[ID]",
			Action: func(c *cli.Context) error {
				idStr := c.Args().Get(0)

				id, err := strconv.Atoi(idStr)
				if err != nil {
					log.Print(err)
					return nil
				}

				return delete(client, int64(id))
				//grpclog.Printf(resp)

				return nil
			},
		},
	}
	app.Run(os.Args)
}
