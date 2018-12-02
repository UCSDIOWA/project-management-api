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

	//Handling Join Requests
	//r, err := c.AddUser(ctx, &pb.AddUserRequest{Projectid: "5c022cc6231ff4000486bd81", Useremail: "test@ucsd.edu"})
	//r, err := c.RemoveUser( context.Background(), &pb.RemoveUserRequest{ Projectid: "5c022cc6231ff4000486bd81", Useremail:"test@ucsd.edu" })
	//r, err := c.RejectUser(ctx, &pb.RejectUserRequest{Projectid: "5c022cc6231ff4000486bd81", Useremail: "test@ucsd.edu"})

	//Handling Invitations
	//r, err := c.InviteUserRequest(ctx, &pb.InviteUserRequest{Projectid:: "test@ucsd.edu", RecipientEmail: "test2@ucsd.edu", SenderEmail: "test@ucsd.edu"})
	//r, err := c.AcceptInvitationRequest(ctx, &pb.AcceptUserRequest{Email: "test@ucsd.edu", Invite: "test2@ucsd.edu invited you to join Easy Like Sunday Mornin' ", Projectid: "5c022cc6231ff4000486bd81"})
	//r, err := c.RejectInvitationRequest(ctx, &pb.RejectUserRequest{Email: "test@ucsd.edu", Invite: "test2@ucsd.edu invited you to join Easy Like Sunday Mornin' ", Projectid: "5c022cc6231ff4000486bd81"})

	//Handling Milestones
	//r, err := c.AddMilestone( context.Background(), &pb.AddMilestoneRequest{ Projectid: "5c022cc6231ff4000486bd81", Title: "123", Description: "test 1", Weight: 1 })
	//r, err := c.EditMilestone( context.Background(), &pb.EditMilestoneRequest{ Milestoneid: "5bff88d87881e70dc4593187", Projectid: "5c022cc6231ff4000486bd81", Title: "New Change", Description:"Changed Description", Weight: 1})
	//r, err := c.MilestoneCompletion( context.Background(), &pb.MilestoneCompletionRequest{ Projectid: "5c022cc6231ff4000486bd81", Milestoneid: "5c01dbf17881e71ad81ecbb9"})
	//r, err := c.DeleteMilestone(context.Background(), &pb.DeleteMilestoneRequest{Projectid: "5c022cc6231ff4000486bd81" , Milestoneid: "5bff88d87881e70dc4593187"})

	//Handling Announcements
	//r, err := c.Announcement(ctx, &pb.Announcement{Projectid: "5c022cc6231ff4000486bd81", Poster: "test@ucsd.edu", Message: "says please finish this", Pin: true})
	//r, err := c.Announcement(ctx, &pb.Announcement{Projectid: "5c022cc6231ff4000486bd81", Poster: "test@ucsd.edu", Message: "says please finish this", Pin: false})

	//Handling TransferLeadership
	//r, err := c.TransferLeader(ctx, &pb.TransferLeaderRequest{Projectid: "5c022cc6231ff4000486bd81", Newleader: "test@ucsd.edu"})

	//Handling GetProjectMembers
	r, err := c.GetProjectMembers(ctx, &pb.GetProjectMembersRequest{Projectid: "5c022cc6231ff4000486bd81", CurrentMembers: []string{"test@ucsd.edu"}})

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
