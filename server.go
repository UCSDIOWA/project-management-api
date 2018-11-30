package main

import (
    "context"
    "log"
    "net"
    "strings"
//    "flag"
//    "os"
//    "net/http"

    pb "github.com/UCSDIOWA/project-management-api/protos"
    "github.com/globalsign/mgo"
    "github.com/globalsign/mgo/bson"
//    "github.com/golang/glog"
    "google.golang.org/grpc"
)

type server struct{}

//Struct to handle current projects of user
type userP struct {
    CurrentProjects []string `json:"currentprojects" bson:"currentprojects"`
}
//Struct to handle users in projects
type projectU struct {
    Users    []string `json:"users" bson:"users"`
}
//Struct to handle project milestones
type projectM struct {
    Milestones []string `json:"milestones" bson:"milestones"`
}
//Struct to handle Milestone Weight
type weightM struct {
    Weight int32 `json:"weight" bson:"weight"`
    Done bool `json:"done" bson:"done"`
}

type mongo struct {
    Operation *mgo.Collection
}

var (
    UserC   *mongo
    ProjC   *mongo
    MileC   *mongo

    //echoEndpoint = flag.String("echo_endpoint", "localhost:50052", "endpoint of project-management-api")
)

func main() {
    errors := make(chan error)

    go func() {
        errors <- startGRPC()
    }()

    /*go func() {
        flag.Parse()
        defer glog.Flush()

        errors <- startHTTP()
    }()*/

    for err := range errors {
        log.Fatal(err)
        return
    }
}

func startGRPC() error {
    //Host mongo server
    m, err := mgo.Dial("127.0.0.1:27017")
    if err != nil {
        log.Fatalf("Could not connecto to the MongoDB server: %v", err)
    }

    defer m.Close()
    log.Println("Connected to MongoDB server")

    //Accessing users collection in tea database
    UserC = &mongo{m.DB("tea").C("users")}
    //Accessing projects collection in tea database
    ProjC = &mongo{m.DB("tea").C("projects")}
    //Accessing milestones collection in tea database
    MileC = &mongo{m.DB("tea").C("milestones")}

    listen, err := net.Listen("tcp", "127.0.0.1:50012")
    if err != nil {
        log.Fatalf("Could not listen on port: %v", err )
    }

    log.Println("Hosting server on", listen.Addr().String())

    s := grpc.NewServer()
    pb.RegisterProjectManagementAPIServer(s, &server{})
    if err := s.Serve(listen); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }

    return err
}

/*func startHTTP() error {
    ctx := context.Background()
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    mux := runtime.NewServeMux()
    opts :=[]grpc.DialOptioin{grpc.WithInsecure()}
    err := pb.RegisterProjectManagementAPIHandlerFromEndPoint(ctx, mux, *echoEndpoint, opts)
    if err != nil {
        return err
    }

    log.Println("Listening on port 8080")

    herokuPort := os.Getenv("PORT")
    if herokuPort == "" {
        herokyPort = "8080"
    }

    return http.ListenAndServe(":"+herokyPort, mux)
}*/


func (s *server) AddMilestone( ctx context.Context, addMileReq *pb.AddMilestoneRequest) (*pb.AddMilestoneResponse, error) {
    xid := bson.NewObjectId().Hex()
    milestone := &pb.MilestoneModel{
        Xid:            xid,
        Title:          addMileReq.Title,
        Description:    addMileReq.Description,
        Users:          addMileReq.Users,
        Weight:         addMileReq.Weight}
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
    milestones.Milestones = append(milestones.Milestones,xid)
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
		"progressbar": finishedWeight*100/totalWeight}}
	err = ProjC.Operation.Update(find, update)
	if err != nil {
		return err
	}

	return nil
}

func (s *server) EditMilestone( ctx context.Context, edMileReq *pb.EditMilestoneRequest) (*pb.EditMilestoneResponse, error) {
    //Find milestone
	find := bson.M{"xid": edMileReq.Milestoneid }
	beforeChange := &weightM{}
	err := MileC.Operation.Find(find).One(beforeChange)
	if err != nil {
		return &pb.EditMilestoneResponse{Success: false}, nil
	}

	//Update Milestone
    milestone := &pb.MilestoneModel{
        Xid:            edMileReq.Milestoneid,
        Title:          edMileReq.Title,
        Description:    edMileReq.Description,
        Users:          edMileReq.Users,
        Weight:         edMileReq.Weight,
    	Done:			beforeChange.Done}

    err = MileC.Operation.Update(find, milestone)
    if err != nil {
        return &pb.EditMilestoneResponse{Success: false}, nil
    }

	//Update progress bar if weight changed
	if( beforeChange.Weight != edMileReq.Weight ) {
		err := updateProgressBar(edMileReq.Projectid)
		if err != nil {
			return &pb.EditMilestoneResponse{Success: false}, nil
		}
	}
    //Otherwise update successful
    return &pb.EditMilestoneResponse{Success: true}, nil
}

func (s *server) MilestoneCompletion( ctx context.Context, milCompReq *pb.MilestoneCompletionRequest) (*pb.MilestoneCompletionResponse, error) {
	//Update milestone status
    find := bson.M{"xid": milCompReq.Milestoneid}
    update := bson.M{"$set": bson.M{"done":true}}
    err := MileC.Operation.Update(find,update)
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

func (s *server) DeleteMilestone( ctx context.Context, rMileReq *pb.DeleteMilestoneRequest) (*pb.DeleteMilestoneResponse, error) {
    return &pb.DeleteMilestoneResponse{}, nil
}

func (s *server) AddUser( ctx context.Context, addUReq *pb.AddUserRequest ) (*pb.AddUserResponse, error) {
    //Fetch User
    userProjects := &userP{}
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
    //Add project to user and update
    userProjects.CurrentProjects = append(userProjects.CurrentProjects, addUReq.Projectid)

    update := bson.M{"$set": bson.M{"currentprojects": userProjects.CurrentProjects}}
    err = UserC.Operation.Update(find, update)
    if err != nil {
        log.Println("User update failed")
        return &pb.AddUserResponse{Success: false}, nil
    }

    //Add user to project and update
    projectUsers.Users = append( projectUsers.Users, addUReq.Useremail )
    update = bson.M{"$set": bson.M{"users": projectUsers.Users}}
    err = ProjC.Operation.Update(findId, update)
    if err != nil {
        log.Println("Projectupdate failed")
        return &pb.AddUserResponse{Success: false}, nil
    }

    //Otherwise Successful
    return &pb.AddUserResponse{Success: true}, nil
}

func (s *server) RemoveUser( ctx context.Context, remUReq *pb.RemoveUserRequest ) (*pb.RemoveUserResponse, error) {
    //Fetch User
    userProjects := &userP{}
    find := bson.M{"email": remUReq.Useremail}

    err := UserC.Operation.Find(find).One(userProjects)
    if err != nil {
        log.Println("Couldn't find user.")
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

    //Find index of project id in user
    for i, num := range userProjects.CurrentProjects {
        if ( strings.Compare(num, remUReq.Projectid) == 0 ){
            userProjects.CurrentProjects[i] = "0"
        }
    }
    //Update user
    update := bson.M{"$set": bson.M{"currentprojects": userProjects.CurrentProjects}}

    err = UserC.Operation.Update(find, update)
    if err != nil {
        return &pb.RemoveUserResponse{Success: false}, nil
    }

    //Find index of user id in project
    for i, num := range projectUsers.Users {
        if( strings.Compare( num, remUReq.Useremail ) == 0 ) {
            projectUsers.Users[i] = "0"
        }
    }
    //Update project
    update = bson.M{"$set": bson.M{"users": projectUsers.Users}}

    err = ProjC.Operation.Update(findId, update)
    if err != nil {
        return &pb.RemoveUserResponse{Success: false}, nil
    }
    //Otherwise success
    return &pb.RemoveUserResponse{Success: true}, nil
}

func (s *server) Announcement( ctx context.Context, annReq *pb.AnnouncementRequest) (*pb.AnnouncementResponse, error) {
    return &pb.AnnouncementResponse{}, nil
}

func (s *server) TransferLeader( ctx context.Context, tlReq *pb.TransferLeaderRequest) (*pb.TransferLeaderResponse, error) {
    return &pb.TransferLeaderResponse{}, nil
}
