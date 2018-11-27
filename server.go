package main

import (
    "context"
    "log"
    "net"
//    "flag"
//    "os"
//    "net/http"

    pb "github.com/UCSDIOWA/project-management-api/protos"
    "github.com/globalsign/mgo"
   // "github.com/globalsign/mgo/bson"
//    "github.com/golang/glog"
    "google.golang.org/grpc"
)

type server struct{}

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

func (s *server) EditProject( ctx context.Context, edProjReq *pb.EditProjectRequest) (*pb.EditProjectResponse, error) {
    //Find project
    project := &pb.ProjectModel{}
    err := ProjC.Operation.FindId(edProjReq.Projectid).One(project)
    if err != nil {
        return &pb.EditProjectResponse{Success: false}, nil
    }
    //Update

    return &pb.EditProjectResponse{Success: true}, nil
}

func (s *server) AddMilestone( ctx context.Context, addMileReq *pb.AddMilestoneRequest) (*pb.AddMilestoneResponse, error) {
    milestone := &pb.MilestoneModel{
        Title:          addMileReq.Title,
        Description:    addMileReq.Description,
        Users:          addMileReq.Users,
        Weight:         addMileReq.Weight}
    err := MileC.Operation.Insert(milestone)
    if err != nil {
        return &pb.AddMilestoneResponse{Success: false}, nil
    }
    return &pb.AddMilestoneResponse{Success: true}, nil
}

func (s *server) EditMilestone( ctx context.Context, edMileReq *pb.EditMilestoneRequest) (*pb.EditMilestoneResponse, error) {
    return &pb.EditMilestoneResponse{}, nil
}

func (s *server) MilestoneCompletion( ctx context.Context, milCompReq *pb.MilestoneCompletionRequest) (*pb.MilestoneCompletionResponse, error) {
    return &pb.MilestoneCompletionResponse{}, nil
}

func (s *server) DeleteMilestone( ctx context.Context, rMileReq *pb.DeleteMilestoneRequest) (*pb.DeleteMilestoneResponse, error) {
    return &pb.DeleteMilestoneResponse{}, nil
}

func (s *server) AddUser( ctx context.Context, addUReq *pb.AddUserRequest ) (*pb.AddUserResponse, error) {
    err := UserC.Operation.FindId(addUReq.Useremail)
    if err != nil {
        return &pb.AddUserResponse{Success: false}, nil
    }

    return &pb.AddUserResponse{Success: true}, nil
}

func (s *server) RemoveUser( ctx context.Context, remUReq *pb.RemoveUserRequest ) (*pb.RemoveUserResponse, error) {
    return &pb.RemoveUserResponse{}, nil
}

func (s *server) Announcement( ctx context.Context, annReq *pb.AnnouncementRequest) (*pb.AnnouncementResponse, error) {
    return &pb.AnnouncementResponse{}, nil
}

func (s *server) TransferLeader( ctx context.Context, tlReq *pb.TransferLeaderRequest) (*pb.TransferLeaderResponse, error) {
    return &pb.TransferLeaderResponse{}, nil
}
