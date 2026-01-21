package pve9x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func tasks() {
	// GET /nodes/{node}/tasks/{upid}/status - Get task status (running)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000001:00000001:00000001:test:running:root@pam:/status$").
		Reply(200).
		JSON(`{
    "data": {
        "status": "running",
        "upid": "UPID:node1:00000001:00000001:00000001:test:running:root@pam:",
        "type": "test",
        "id": "running",
        "user": "root@pam",
        "node": "node1",
        "pid": 1,
        "pstart": 1,
        "starttime": 1693252591
    }
}`)

	// GET /nodes/{node}/tasks/{upid}/status - Get task status (stopped/completed)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000002:00000002:00000002:test:completed:root@pam:/status$").
		Reply(200).
		JSON(`{
    "data": {
        "status": "stopped",
        "exitstatus": "OK",
        "upid": "UPID:node1:00000002:00000002:00000002:test:completed:root@pam:",
        "type": "test",
        "id": "completed",
        "user": "root@pam",
        "node": "node1",
        "pid": 2,
        "pstart": 2,
        "starttime": 1693252591,
        "endtime": 1693252600
    }
}`)

	// GET /nodes/{node}/tasks/{upid}/status - Get task status (failed)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000003:00000003:00000003:test:failed:root@pam:/status$").
		Reply(200).
		JSON(`{
    "data": {
        "status": "stopped",
        "exitstatus": "some error occurred",
        "upid": "UPID:node1:00000003:00000003:00000003:test:failed:root@pam:",
        "type": "test",
        "id": "failed",
        "user": "root@pam",
        "node": "node1",
        "pid": 3,
        "pstart": 3,
        "starttime": 1693252591,
        "endtime": 1693252600
    }
}`)

	// DELETE /nodes/{node}/tasks/{upid} - Stop task
	gock.New(config.C.URI).
		Delete("^/nodes/node1/tasks/UPID:node1:00000001:00000001:00000001:test:running:root@pam:$").
		Reply(200).
		JSON(`{
    "data": null
}`)

	// GET /nodes/{node}/tasks/{upid}/log - Get task log
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000002:00000002:00000002:test:completed:root@pam:/log$").
		MatchParam("start", "0").
		MatchParam("limit", "50").
		Reply(200).
		JSON(`{
    "data": [
        {"n": 1, "t": "task started"},
        {"n": 2, "t": "processing step 1"},
        {"n": 3, "t": "processing step 2"},
        {"n": 4, "t": "processing step 3"},
        {"n": 5, "t": "task completed successfully"}
    ]
}`)

	// GET /nodes/{node}/tasks/{upid}/log - Get task log with offset
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000002:00000002:00000002:test:completed:root@pam:/log$").
		MatchParam("start", "5").
		MatchParam("limit", "50").
		Reply(200).
		JSON(`{
    "data": []
}`)

	// GET /nodes/{node}/tasks/{upid}/log - Get task log for running task
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000001:00000001:00000001:test:running:root@pam:/log$").
		Reply(200).
		JSON(`{
    "data": [
        {"n": 1, "t": "task started"},
        {"n": 2, "t": "processing..."}
    ]
}`)
}
