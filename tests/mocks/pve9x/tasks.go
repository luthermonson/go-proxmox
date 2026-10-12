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

	// GET /nodes/{node}/tasks/{upid} — per-task diridx (see nodes_diridx.go).
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000002:00000002:00000002:test:completed:root@pam:$").
		Reply(200).
		JSON(`{"data":[
			{"subdir":"log"},
			{"subdir":"status"}
		]}`)

	// --- Watch / Wait test fixtures ----------------------------------------
	// status — task already stopped, so Wait/Watch terminate on first ping.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000099:00000099:00000099:watchme:99:root@pam:/status$").
		Reply(200).
		JSON(`{
    "data": {
        "status": "stopped",
        "exitstatus": "OK",
        "upid": "UPID:node1:00000099:00000099:00000099:watchme:99:root@pam:",
        "type": "watchme",
        "id": "99",
        "user": "root@pam",
        "node": "node1",
        "starttime": 1700000000,
        "endtime": 1700000005
    }
}`)
	// log — has entries so Watch returns a channel.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000099:00000099:00000099:watchme:99:root@pam:/log$").
		Reply(200).
		JSON(`{"data":[
			{"n":1,"t":"line one"},
			{"n":2,"t":"line two"}
		]}`)

	// status — task that never produces logs (used to exercise the
	// "no logs available" error path in Watch).
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:000000AA:000000AA:000000AA:emptylog:AA:root@pam:/status$").
		Reply(200).
		JSON(`{"data":{"status":"stopped","exitstatus":"OK","upid":"UPID:node1:000000AA:000000AA:000000AA:emptylog:AA:root@pam:","type":"emptylog","id":"AA","user":"root@pam","node":"node1"}}`)
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:000000AA:000000AA:000000AA:emptylog:AA:root@pam:/log$").
		Reply(200).
		JSON(`{"data":[]}`)

	// status — task that succeeded with warnings. PVE reports this as a
	// stopped task whose exitstatus is "WARNINGS: <n>".
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/node1/tasks/UPID:node1:00000004:00000004:00000004:test:warnings:root@pam:/status$").
		Reply(200).
		JSON(`{"data":{"status":"stopped","exitstatus":"WARNINGS: 2","upid":"UPID:node1:00000004:00000004:00000004:test:warnings:root@pam:","type":"test","id":"warnings","user":"root@pam","node":"node1","starttime":1693252591,"endtime":1693252600}}`)

	// --- "no such task" fallback fixtures ----------------------------------
	// Node taskhist has its own task index so these don't collide with the
	// persisted ^/nodes/node1/tasks$ list in node_loose.go. Each status
	// endpoint answers the way PVE does once a task's status file is gone:
	// a 400 parameter error {"upid":"no such task"}.
	for _, upid := range []string{
		"UPID:taskhist:0000A001:0000A001:65000000:qmshutdown:100:root@pam:",
		"UPID:taskhist:0000A002:0000A002:65000010:qmstart:100:root@pam:",
		"UPID:taskhist:0000A003:0000A003:65000020:qmstart:101:root@pam:",
	} {
		gock.New(config.C.URI).
			Persist().
			Get("^/nodes/taskhist/tasks/" + upid + "/status$").
			Reply(400).
			JSON(`{"data":null,"errors":{"upid":"no such task"}}`)
	}
	// The node task index still has the first two (one failed, one OK); the
	// third has aged out of the history too.
	gock.New(config.C.URI).
		Persist().
		Get("^/nodes/taskhist/tasks$").
		MatchParam("source", "all").
		Reply(200).
		JSON(`{"data":[
			{
				"upid":"UPID:taskhist:0000A001:0000A001:65000000:qmshutdown:100:root@pam:",
				"node":"taskhist","pid":40961,"pstart":40961,"starttime":1694498816,
				"endtime":1694498996,"type":"qmshutdown","id":"100","user":"root@pam",
				"status":"VM quit/powerdown failed - got timeout"
			},
			{
				"upid":"UPID:taskhist:0000A002:0000A002:65000010:qmstart:100:root@pam:",
				"node":"taskhist","pid":40962,"pstart":40962,"starttime":1694498832,
				"endtime":1694498840,"type":"qmstart","id":"100","user":"root@pam",
				"status":"OK"
			}
		]}`)
}
