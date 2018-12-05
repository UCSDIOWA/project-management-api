package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	pb "github.com/UCSDIOWA/project-management-api/protos"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/rs/cors"

	"google.golang.org/grpc"
)

type server struct{}

//Struct to handle user collection
type user struct {
	Invitations     []string `json:"invitations" bson:"invitations"`
	ProjectInvites  []string `json:"projectinvites" bson:"projectinvites"`
	Notifications   []string `json:"notifications" bson:"notifications"`
	CurrentProjects []string `json:"currentprojects" bson:"currentprojects"`
}

//Struct to handle users in projects
type projectU struct {
	Title string   `json:"title" bson:"title"`
	Users []string `json:"memberslist" bson:"memberslist"`
}

//Struct to handle pending users in projects
type userJoinReqs struct {
	JoinRequests []string `json:"joinrequests" bson:"joinrequests"`
}

//Struct to handle project milestones
type projectM struct {
	Milestones []string `json:"milestones" bson:"milestones"`
}

//Struct to handle user invitations
type projectT struct {
	Title string `json:"title" bson:"title"`
}

//Struct to handle user invitations
type projectA struct {
	PinnedAnnouncements   []string `json:"pinnedannouncements" bson:"pinnedannouncements"`
	UnpinnedAnnouncements []string `json:"unpinnedannouncements" bson:"unpinnedannouncements"`
}

//Struct to handle Milestone Weight
type weightM struct {
	Weight int32 `json:"weight" bson:"weight"`
	Done   bool  `json:"done" bson:"done"`
}

type mongo struct {
	Operation *mgo.Collection
}

var (
	UserC *mongo
	ProjC *mongo
	MileC *mongo

	echoEndpoint = flag.String("echo_endpoint", "localhost:50052", "endpoint of project-management-api")
)

func main() {
	errors := make(chan error)

	go func() {
		errors <- startGRPC()
	}()

	go func() {
		flag.Parse()
		defer glog.Flush()

		errors <- startHTTP()
	}()

	for err := range errors {
		log.Fatal(err)
		return
	}
}

func startGRPC() error {
	// Host mongo server
	m, err := mgo.Dial("127.0.0.1:27017")
	if err != nil {
		log.Fatalf("Could not connect to the MongoDB server: %v", err)
	}
	defer m.Close()
	log.Println("Connected to MongoDB server")

	UserC = &mongo{m.DB("tea").C("users")}
	ProjC = &mongo{m.DB("tea").C("projects")}
	MileC = &mongo{m.DB("tea").C("milestones")}

	// Host grpc server
	listen, err := net.Listen("tcp", "127.0.0.1:50052")
	if err != nil {
		log.Fatalf("Could not listen on port: %v", err)
	}

	log.Println("Hosting server on", listen.Addr().String())

	s := grpc.NewServer()
	pb.RegisterProjectManagementAPIServer(s, &server{})
	if err := s.Serve(listen); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	return err
}

func startHTTP() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gwmux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := pb.RegisterProjectManagementAPIHandlerFromEndpoint(ctx, gwmux, *echoEndpoint, opts)
	if err != nil {
		return err
	}

	log.Println("Listening on port 8080")

	mux := http.NewServeMux()
	mux.HandleFunc("/.*", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST")
	})
	mux.Handle("/", gwmux)
	handler := cors.Default().Handler(mux)

	herokuPort := os.Getenv("PORT")
	if herokuPort == "" {
		herokuPort = "8080"
	}

	return http.ListenAndServe(":"+herokuPort, handler)
}

func (s *server) AddMilestone(ctx context.Context, addMileReq *pb.AddMilestoneRequest) (*pb.AddMilestoneResponse, error) {
	xid := bson.NewObjectId().Hex()
	milestone := &pb.MilestoneModel{
		Xid:         xid,
		Title:       addMileReq.Title,
		Description: addMileReq.Description,
		Users:       addMileReq.Users,
		Weight:      addMileReq.Weight}
	err := MileC.Operation.Insert(milestone)
	if err != nil {
		return &pb.AddMilestoneResponse{Success: false}, nil
	}
	//Add to project milestones
	milestones := &projectM{}
	find := bson.M{"xid": addMileReq.Projectid}

	err = ProjC.Operation.Find(find).One(milestones)
	if err != nil {
		return &pb.AddMilestoneResponse{Success: false}, nil
	}

	//Update project
	milestones.Milestones = append(milestones.Milestones, xid)
	update := bson.M{"$set": bson.M{"milestones": milestones.Milestones}}
	err = ProjC.Operation.Update(find, update)
	if err != nil {
		return &pb.AddMilestoneResponse{Success: false}, nil
	}
	//Update progress bar
	err = updateProgressBar(addMileReq.Projectid)
	if err != nil {
		return &pb.AddMilestoneResponse{Success: false}, nil
	}
	//Otherwise everything is good
	return &pb.AddMilestoneResponse{Success: true}, nil
}

func updateProgressBar(projectid string) error {
	milestones := &projectM{}
	find := bson.M{"xid": projectid}
	err := ProjC.Operation.Find(find).One(milestones)
	if err != nil {
		return err
	}
	var finishedWeight int32
	var totalWeight int32
	finishedWeight = 0
	totalWeight = 0
	currentMile := &weightM{}
	for _, cur := range milestones.Milestones {
		if strings.Compare(cur, "0") != 0 {
			findMile := bson.M{"xid": cur}
			err := MileC.Operation.Find(findMile).One(currentMile)
			if err != nil {
				return err
			}
			totalWeight += currentMile.Weight
			if currentMile.Done {
				finishedWeight += currentMile.Weight
			}
		}
	}

	update := bson.M{"$set": bson.M{"milestones": milestones.Milestones,
		"progressbar": finishedWeight * 100 / totalWeight}}
	err = ProjC.Operation.Update(find, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *server) EditMilestone(ctx context.Context, edMileReq *pb.EditMilestoneRequest) (*pb.EditMilestoneResponse, error) {
	//Find milestone
	find := bson.M{"xid": edMileReq.Milestoneid}
	beforeChange := &weightM{}
	err := MileC.Operation.Find(find).One(beforeChange)
	if err != nil {
		return &pb.EditMilestoneResponse{Success: false}, nil
	}

	//Update Milestone
	milestone := &pb.MilestoneModel{
		Xid:         edMileReq.Milestoneid,
		Title:       edMileReq.Title,
		Description: edMileReq.Description,
		Users:       edMileReq.Users,
		Weight:      edMileReq.Weight,
		Done:        beforeChange.Done}

	err = MileC.Operation.Update(find, milestone)
	if err != nil {
		return &pb.EditMilestoneResponse{Success: false}, nil
	}

	//Update progress bar if weight changed
	if beforeChange.Weight != edMileReq.Weight {
		err := updateProgressBar(edMileReq.Projectid)
		if err != nil {
			return &pb.EditMilestoneResponse{Success: false}, nil
		}
	}
	//Otherwise update successful
	return &pb.EditMilestoneResponse{Success: true}, nil
}

func (s *server) MilestoneCompletion(ctx context.Context, milCompReq *pb.MilestoneCompletionRequest) (*pb.MilestoneCompletionResponse, error) {
	//Update milestone status
	find := bson.M{"xid": milCompReq.Milestoneid}
	update := bson.M{"$set": bson.M{"done": true}}
	err := MileC.Operation.Update(find, update)
	if err != nil {
		return &pb.MilestoneCompletionResponse{Success: false}, nil
	}
	//Update progress bar
	err = updateProgressBar(milCompReq.Projectid)
	if err != nil {
		return &pb.MilestoneCompletionResponse{Success: false}, nil
	}
	return &pb.MilestoneCompletionResponse{Success: true}, nil
}

func (s *server) DeleteMilestone(ctx context.Context, delMileReq *pb.DeleteMilestoneRequest) (*pb.DeleteMilestoneResponse, error) {
	//Get Project milestones
	milestones := &projectM{}
	find := bson.M{"xid": delMileReq.Projectid}
	err := ProjC.Operation.Find(find).One(milestones)
	if err != nil {
		return &pb.DeleteMilestoneResponse{Success: false}, nil
	}
	//Find Milestone and delete
	for i, cur := range milestones.Milestones {
		if strings.Compare(cur, delMileReq.Milestoneid) == 0 {
			milestones.Milestones[i] = "0"
		}
	}
	//Update database
	update := bson.M{"$set": bson.M{"milestones": milestones.Milestones}}
	err = ProjC.Operation.Update(find, update)
	if err != nil {
		return &pb.DeleteMilestoneResponse{Success: false}, nil
	}

	//Update Progressbar
	err = updateProgressBar(delMileReq.Projectid)
	if err != nil {
		return &pb.DeleteMilestoneResponse{Success: false}, nil
	}
	//Otherwise everything is good
	return &pb.DeleteMilestoneResponse{Success: true}, nil
}

func (s *server) AddUser(ctx context.Context, addUReq *pb.AddUserRequest) (*pb.AddUserResponse, error) {
	//Fetch User
	userProjects := &user{}
	find := bson.M{"email": addUReq.Useremail}

	err := UserC.Operation.Find(find).One(userProjects)
	if err != nil {
		log.Println("Couldn't find user.")
		return &pb.AddUserResponse{Success: false}, nil
	}

	//Fetch project
	projectUsers := &projectU{}
	findId := bson.M{"xid": addUReq.Projectid}
	err = ProjC.Operation.Find(findId).One(projectUsers)
	if err != nil {
		log.Println("Couldn't find project.")
		return &pb.AddUserResponse{Success: false}, nil
	}

	//Add project to user and update current projects and notifications
	userProjects.CurrentProjects = append(userProjects.CurrentProjects, addUReq.Projectid)
	userProjects.Notifications = append([]string{"You've been added to the project " + projectUsers.Title}, userProjects.Notifications...)
	update := bson.M{"$set": bson.M{"currentprojects": userProjects.CurrentProjects, "notifications": userProjects.Notifications}}
	err = UserC.Operation.Update(find, update)
	if err != nil {
		log.Println("User update failed")
		return &pb.AddUserResponse{Success: false}, nil
	}

	//send notifications to all of the members of the group
	for _, usr := range projectUsers.Users {
		userProjects := &user{}
		find = bson.M{"email": usr}
		err = UserC.Operation.Find(find).One(userProjects)
		if err != nil {
			log.Println("Couldn't find user.")
			return &pb.AddUserResponse{Success: false}, nil
		}
		userProjects.Notifications = append(userProjects.Notifications, addUReq.Useremail+" has been added to the project "+projectUsers.Title)
		err = UserC.Operation.Update(bson.M{"email": usr}, bson.M{"notifications": userProjects.Notifications})
		if err != nil {
			log.Println("User update failed")
			return &pb.AddUserResponse{Success: false}, nil
		}
	}

	//Add user to project and update
	projectUsers.Users = append(projectUsers.Users, addUReq.Useremail)
	update = bson.M{"$set": bson.M{"memberslist": projectUsers.Users}}
	err = ProjC.Operation.Update(findId, update)
	if err != nil {
		log.Println("Projectupdate failed")
		return &pb.AddUserResponse{Success: false}, nil
	}

	//Otherwise Successful
	return &pb.AddUserResponse{Success: true}, nil
}

func (s *server) RemoveUser(ctx context.Context, remUReq *pb.RemoveUserRequest) (*pb.RemoveUserResponse, error) {
	//Fetch User
	userProjects := &user{}
	find := bson.M{"email": remUReq.Useremail}
	err := UserC.Operation.Find(find).One(userProjects)
	if err != nil {
		log.Println("Couldn't find user.")
		return &pb.RemoveUserResponse{Success: false}, nil
	}

	//Find index of project id in user, remove the user
	for i, num := range userProjects.CurrentProjects {
		if strings.Compare(num, remUReq.Projectid) == 0 {
			userProjects.CurrentProjects[i] = userProjects.CurrentProjects[len(userProjects.CurrentProjects)-1]
			userProjects.CurrentProjects = userProjects.CurrentProjects[:len(userProjects.CurrentProjects)-1]
		}
	}

	//Update user
	update := bson.M{"$set": bson.M{"currentprojects": userProjects.CurrentProjects}}
	err = UserC.Operation.Update(find, update)
	if err != nil {
		return &pb.RemoveUserResponse{Success: false}, nil
	}

	//Fetch project
	projectUsers := &projectU{}
	findId := bson.M{"xid": remUReq.Projectid}
	err = ProjC.Operation.Find(findId).One(projectUsers)
	if err != nil {
		log.Println("Couldn't find project.")
		return &pb.RemoveUserResponse{Success: false}, nil
	}

	//Find index of user id in project
	for i, num := range projectUsers.Users {
		if strings.Compare(num, remUReq.Useremail) == 0 {
			projectUsers.Users[i] = projectUsers.Users[len(projectUsers.Users)]
		}
	}
	//Update project
	update = bson.M{"$set": bson.M{"memberslist": projectUsers.Users}}

	err = ProjC.Operation.Update(findId, update)
	if err != nil {
		return &pb.RemoveUserResponse{Success: false}, nil
	}
	//Otherwise success
	return &pb.RemoveUserResponse{Success: true}, nil
}

func (s *server) RejectUser(ctx context.Context, rejUsrReq *pb.RejectUserRequest) (*pb.RejectUserResponse, error) {

	//Fetch project
	pendingUsers := &userJoinReqs{}
	findId := bson.M{"xid": rejUsrReq.Projectid}

	err := ProjC.Operation.Find(findId).One(pendingUsers)
	if err != nil {
		log.Println("Couldn't find project.")
		return &pb.RejectUserResponse{Success: false}, nil
	}

	//Remove user from pending users in project and update
	for i, usr := range pendingUsers.JoinRequests {
		if usr == rejUsrReq.Useremail {
			pendingUsers.JoinRequests[i] = pendingUsers.JoinRequests[len(pendingUsers.JoinRequests)-1]
			pendingUsers.JoinRequests = pendingUsers.JoinRequests[:len(pendingUsers.JoinRequests)-1]
			break
		}
	}
	update := bson.M{"$set": bson.M{"joinrequests": pendingUsers.JoinRequests}}
	err = ProjC.Operation.Update(findId, update)
	if err != nil {
		log.Println("Project update with new Pending-Users failed")
		return &pb.RejectUserResponse{Success: false}, nil
	}

	//Otherwise Successful
	return &pb.RejectUserResponse{Success: true}, nil
}

func (s *server) GetProjectMembers(ctx context.Context, currMembs *pb.GetProjectMembersRequest) (*pb.GetProjectMembersResponse, error) {

	//for each email, find the user of that email and get it's first name and email
	users := []*pb.UserTuple{}
	for _, currEmail := range currMembs.CurrentMembers {
		//get this user's email and first name
		userInfo := &pb.UserTuple{}
		findID := bson.M{"email": currEmail}
		err := UserC.Operation.Find(findID).One(userInfo)
		if err != nil {
			log.Println("Finding user based on given email failed")
			return &pb.GetProjectMembersResponse{Success: false}, nil
		}
		if userInfo.Email == "" || userInfo.Firstname == "" {
			log.Println("Failed to retrieve user's first name and email")
			return &pb.GetProjectMembersResponse{Success: false}, nil
		}
		//append this user's first name and email to our array of tuples
		users = append(users, userInfo)
	}

	//return success and the array of tuples
	return &pb.GetProjectMembersResponse{Success: true, Users: users}, nil
}

func (s *server) InviteUser(ctx context.Context, invite *pb.InviteUserRequest) (*pb.InviteUserResponse, error) {

	//get user's invitations
	invites := user{}
	findID := bson.M{"email": invite.Receipientemail}
	err := UserC.Operation.Find(findID).One(invites)
	if err != nil {
		log.Println("Finding user based on given email failed")
		return &pb.InviteUserResponse{Success: false}, nil
	}

	//get the project based on the project id
	projTitle := projectT{}
	find := bson.M{"xid": invite.Projectid}
	err = ProjC.Operation.Find(find).One(projTitle)
	if err != nil {
		log.Println("Finding project based on given xid failed")
		return &pb.InviteUserResponse{Success: false}, nil
	}

	//add the new invitation and notification, and update the database
	invites.Invitations = append(invites.Invitations, invite.Senderemail+"invited you to join "+projTitle.Title)
	invites.ProjectInvites = append(invites.ProjectInvites, invite.Projectid)
	invites.Notifications = append([]string{invite.Senderemail + "invited you to join " + projTitle.Title}, invites.Notifications...)
	err = UserC.Operation.Update(findID, bson.M{"invitations": invites.Invitations, "projectinvites": invites.ProjectInvites})
	if err != nil {
		log.Println("Updating user's invitations failed")
		return &pb.InviteUserResponse{Success: false}, nil
	}

	//return success
	return &pb.InviteUserResponse{Success: true}, nil

}

func (s *server) RejectInvitation(ctx context.Context, invite *pb.RejectInviteRequest) (*pb.RejectInviteResponse, error) {

	//get user's invitations
	invites := user{}
	findID := bson.M{"email": invite.Email}
	err := UserC.Operation.Find(findID).One(invites)
	if err != nil {
		log.Println("Finding user based on given email failed")
		return &pb.RejectInviteResponse{Success: false}, nil
	}

	//add the new invitation
	for i, currInvite := range invites.Invitations {
		//check if the email and the title are in this invitation, and remove it once found
		if currInvite == invites.Invitations[i] {
			invites.Invitations[i] = invites.Invitations[len(invites.Invitations)-1]
			invites.Invitations = invites.Invitations[:len(invites.Invitations)-1]
			break
		}
	}

	//update the database
	err = UserC.Operation.Update(findID, bson.M{"invitations": invites.Invitations})
	if err != nil {
		log.Println("Updating user's invitations failed")
		return &pb.RejectInviteResponse{Success: false}, nil
	}

	//return success
	return &pb.RejectInviteResponse{Success: true}, nil

}

func (s *server) AcceptInvitation(ctx context.Context, invite *pb.AcceptInviteRequest) (*pb.AcceptInviteResponse, error) {

	//Fetch project
	projectUsers := &projectU{}
	find := bson.M{"xid": invite.Projectid}
	err := ProjC.Operation.Find(find).One(projectUsers)
	if err != nil {
		log.Println("Couldn't find project.")
		return &pb.AcceptInviteResponse{Success: false}, nil
	}

	//Add project to user and update
	projectUsers.Users = append(projectUsers.Users, invite.Email)
	update := bson.M{"$set": bson.M{"memberslist": projectUsers.Users}}
	err = ProjC.Operation.Update(find, update)
	if err != nil {
		log.Println("User update failed")
		return &pb.AcceptInviteResponse{Success: false}, nil
	}

	//get user's invitations
	invites := user{}
	findID := bson.M{"email": invite.Email}
	err = UserC.Operation.Find(findID).One(invites)
	if err != nil {
		log.Println("Finding user based on given email failed")
		return &pb.AcceptInviteResponse{Success: false}, nil
	}

	//add the new invitation
	for i, currInvite := range invites.Invitations {
		//check if the email and the title are in this invitation, and remove it once found
		if currInvite == invites.Invitations[i] {
			invites.Invitations[i] = invites.Invitations[len(invites.Invitations)-1]
			invites.Invitations = invites.Invitations[:len(invites.Invitations)-1]
			break
		}
	}

	//update the database
	err = UserC.Operation.Update(findID, bson.M{"invitations": invites.Invitations})
	if err != nil {
		log.Println("Updating user's invitations failed")
		return &pb.AcceptInviteResponse{Success: false}, nil
	}

	//return success
	return &pb.AcceptInviteResponse{Success: true}, nil

}

func (s *server) Announcement(ctx context.Context, annReq *pb.AnnouncementRequest) (*pb.AnnouncementResponse, error) {

	//retrieve project
	oldAnnouncements := &projectA{}
	find := bson.M{"xid": annReq.Projectid}
	err := ProjC.Operation.Find(find).One(oldAnnouncements)
	if err != nil {
		log.Println("Finding project based on given xid failed")
		return &pb.AnnouncementResponse{Success: false}, nil
	}
	//determine whether to add the post to the top or the bottom of the announcements
	if annReq.Pin {
		oldAnnouncements.PinnedAnnouncements = append([]string{annReq.Poster + "says " + annReq.Message}, oldAnnouncements.PinnedAnnouncements...)
	} else {
		oldAnnouncements.UnpinnedAnnouncements = append([]string{annReq.Poster + "says " + annReq.Message}, oldAnnouncements.UnpinnedAnnouncements...)
	}

	//update the database
	err = ProjC.Operation.Update(find, bson.M{"pinnedannouncements": oldAnnouncements.PinnedAnnouncements,
		"unpinnedannouncements": oldAnnouncements.UnpinnedAnnouncements})
	if err != nil {
		log.Println("Updating projects invitations failed")
		return &pb.AnnouncementResponse{Success: false}, nil
	}

	return &pb.AnnouncementResponse{Success: true}, nil
}

func (s *server) TransferLeader(ctx context.Context, tlReq *pb.TransferLeaderRequest) (*pb.TransferLeaderResponse, error) {

	//update the leadership
	findID := bson.M{"xid:": tlReq.Projectid}
	err := ProjC.Operation.Update(findID, bson.M{"projectleader": tlReq.Newleader})
	if err != nil {
		log.Println("Finding project based on given xid failed")
		return &pb.TransferLeaderResponse{Success: false}, nil
	}

	return &pb.TransferLeaderResponse{Success: true}, nil
}

func (s *server) RemoveNotification(ctx context.Context, nReq *pb.RemoveNotificationRequest) (*pb.RemoveNotificationResponse, error) {

	//remove the notification
	userNotifications := user{}
	findID := bson.M{"email:": nReq.User}
	for i, notif := range userNotifications.Notifications {
		//check when we find notification
		if notif == nReq.Notification {
			userNotifications.Notifications[i] = (userNotifications.Notifications[len(userNotifications.Notifications)-1])
			userNotifications.Notifications = userNotifications.Notifications[:len(userNotifications.Notifications)-1]
		}
	}
	err := UserC.Operation.Update(findID, bson.M{"notifications": userNotifications.Notifications})
	if err != nil {
		log.Println("Finding user and updating their notifications based on given email failed")
		return &pb.RemoveNotificationResponse{Success: false}, nil
	}

	return &pb.RemoveNotificationResponse{Success: true}, nil
}
