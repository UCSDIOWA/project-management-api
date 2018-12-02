package main

import (
	"context"
	"log"
	"strconv"
	"time"

	pb "github.com/UCSDIOWA/project-management-api/protos"
	"google.golang.org/grpc"
)

func main() {

	// Connect to the server
	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect to the server: %v", err)
	}
	defer conn.Close()
	c := pb.NewProjectManagementAPIClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	r, err := c.AddUser(ctx, &pb.AddUserRequest{Projectid: "5c022cc6231ff4000486bd81", Useremail: "test@ucsd.edu"})
	if err != nil {
		log.Println("An err occurred.")
		log.Println(err)
	} else {
		log.Println(r.String())
		log.Println("Success: " + strconv.FormatBool(r.Success))
	}
	//r, err := c.RemoveUser( context.Background(), &pb.RemoveUserRequest{ Projectid: "5bfdbf4aa9b8c54a90fdedbf", Useremail:"test@ucsd.edu" })
	//r, err := c.AddMilestone( context.Background(), &pb.AddMilestoneRequest{ Projectid: "5bfdbf4aa9b8c54a90fdedbf", Title: "123", Description: "test 1", Weight: 1 })
	//r, err := c.EditMilestone( context.Background(), &pb.EditMilestoneRequest{ Milestoneid: "5bff58f37881e71060e29f26", Projectid: "5bfdbf4aa9b8c54a90fdedbf", Title: "New Change", Description:"Changed Description", Weight: 1})
	//r, err := c.MilestoneCompletion( context.Background(), &pb.MilestoneCompletionRequest{ Projectid: "5bfdbf4aa9b8c54a90fdedbf", Milestoneid: "5c01dbf17881e71ad81ecbb9"})
	//r, err := c.DeleteMilestone(context.Background(), &pb.DeleteMilestoneRequest{Projectid: "5bfdbf4aa9b8c54a90fdedbf" , Milestoneid: "5bff88d87881e70dc4593187"})
}
