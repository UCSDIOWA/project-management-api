This repository contains the necessary code for all endpoints pertaining to project management. This means anything that can occur from the project dashboard page is handled by the endpoints here. All database collections are modified from this API, although not necessarily by every endpoint. The endpoints defined in this repository are below, along with the POST Request and Response bodies as Protocol Buffers. The gRPC server runs on a server hosted on Heroku, as does the MongoDB database which stores all of our data. vendor and GoDeps are directories necessary for the server to run properly on Heroku. The protos directory contains the protocol buffer definitions and the corresponding generated pb.gw and pb Go files. 


Endpoint: /addmilestone

Handles adding a milestone to the project.

message AddMilestoneRequest {
    string xid = 1;
    string title = 2;
    string description = 3;
    repeated string users = 4;
    int32 weight = 5;
}

message AddMilestoneResponse {
    bool success = 1;
}



Endpoint: /editmilestone

Handles editing the details of a pre-existing milestone that is a part of the project.

message EditMilestoneRequest {
    string milestoneid = 1;
    string title = 2;
    string description = 3;
    repeated string users = 4;
    int32 weight = 5;
}

message EditMilestoneResponse {
    bool success = 1;
}



Endpoint: /deletemilestone

Handles removing a milestone from the project.

message DeleteMilestoneRequest {
    string xid = 1;
    string milestoneid = 2;                                             
}

message DeleteMilestoneResponse {
    bool success = 1;
}



Endpoint: /milestonecompletion

Handles toggling the milestone as complete or incomplete.

message MilestoneCompletionRequest {
    string milestoneid = 2;
}

message MilestoneCompletionResponse {
    bool success = 1;
}



Endpoint: /getallmilestones

Helper function for the Front-End to retrieve all Milestones pertaining to a project. The array of MilestoneModels that is returned is described below as well.

message GetAllMilestonesRequest {
    repeated string milestoneid = 1;
}

message GetAllMilestonesResponse {
    bool success = 1;
    repeated MilestoneModel milestones = 2;
}

message MilestoneModel { 
    string milestoneid = 1;
    string title = 2;
    string description = 3;
    repeated string users =4;
    int32 weight = 5;
    bool done = 6;
}



Endpoint: /getprojectmembers

Helper function for the Front-End to retrieve certain user information for the users who are a part of the project. The array of UserTuples that is returned is described below as well.

message GetProjectMembersRequest {
    string xid = 1;
    repeated string memberslist = 2;
}

message GetProjectMembersResponse {
    bool success = 1;
    repeated UserTuple users = 2;
}

message UserTuple {
    string email = 1;
    string firstname = 2;
}



Endpoint: /inviteuser

Handles the case when someone who is a part of the project invites a user who is not a part of the project to join the project.

message InviteUserRequest {
    string xid = 1;
    string recipientemail = 2;
    string senderemail = 3;
}

message InviteUserResponse {
    bool success = 1;
}



Endpoint: /acceptinvitation

Handles the case when a user accepts an invitation to join a project. The invitation is removed from their invitations list.

message AcceptInviteRequest {
    string email = 1;
    string xid = 2;
}

message AcceptInviteResponse {
    bool success = 1;
}



Endpoint: /rejectinvitation

Handles the case when a user rejects an invitation to join a project. The invitation is removed from their invitations list.

message RejectInviteRequest {
    string email = 1;
    string xid = 2;
}

message RejectInviteResponse {
    bool success = 1;
}



Endpoint: /adduser

Adds a user who has requested to join the project to the project. The user's request is removed from the requests array and their email is added to the members list array.

message AddUserRequest {
    string xid = 1;
    string email = 2;
}

message AddUserResponse {
    bool success = 1;
}



Endpoint: /removeuser

Removes a user who is presently a part of the project.

message RemoveUserRequest {
    string xid = 1;
    string email = 2;
}

message RemoveUserResponse {
    bool success = 1;
}



Endpoint: /rejectuser

Rejects a user's request to join the project, and removes their request from the join requests array.

message RejectUserRequest {
    string xid = 1;
    string email = 2;
}

message RejectUserResponse {
    bool success = 1;
}



Endpoint: /transferleader

Changes the leader from one person in the project to another person in the group.

message TransferLeaderRequest {
    string xid = 1;
    string newleader = 2;
}

message TransferLeaderResponse {
    bool success = 1;
}



Endpoint: /announcement

Makes either a pinned or unpinned announcement as a part of the project.

message AnnouncementRequest {
    string xid = 1;
    string poster = 2;
    string message = 3;
    bool pin = 4;
}

message AnnouncementResponse {
    bool success = 1;
}



Endpoint: /removenotification

Removes a notification for a user.

message RemoveNotificationRequest {
    string notification = 1;
    string user = 2;
}

message RemoveNotificationResponse{
    bool success = 1;
}


Author: Samuel Blake
Last Updated: 12/6/2018
