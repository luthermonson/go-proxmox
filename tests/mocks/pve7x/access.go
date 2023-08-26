package pve7x

import (
	"github.com/h2non/gock"
	"github.com/luthermonson/go-proxmox/tests/mocks/config"
)

func access() {
	gock.New(config.C.URI).
		Get("^/access/domains$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "type": "pve",
            "realm": "pve",
            "comment": "Proxmox VE authentication server"
        },
        {
            "type": "pam",
            "realm": "pam",
            "comment": "Linux PAM standard authentication"
        },
        {
            "realm": "test",
            "type": "ldap",
            "tfa": "oath",
            "comment": "comment comment comment"
        }
    ]
}`)
	gock.New(config.C.URI).
		Get("^/access/domains/test$").
		Reply(200).
		JSON(`{
    "data": {
        "user_attr": "userattribute",
        "sync-defaults-options": "remove-vanished=acl;entry;properties,scope=users",
        "port": 1234,
        "server1": "server1",
        "user_classes": "userclasses",
        "tfa": "digits=8,step=1234,type=oath",
        "comment": "comment comment comment",
        "group_name_attr": "groupnameattr",
        "digest": "b84e9112ebbb173fc8f5af76a057b38178f1047c",
        "secure": 1,
        "default": 0,
        "sync_attributes": "email=email@attribute.com",
        "base_dn": "CN=Users",
        "type": "ldap",
        "bind_dn": "CN=Users",
        "group_filter": "groupfilter",
        "group_classes": "groupclasses",
        "verify": 1,
        "filter": "userfilter",
        "server2": "server2"
    }
}`)

	gock.New(config.C.URI).
		Get("^/access$").
		Reply(200).
		JSON(`
{
    "data": [
        {
            "subdir": "users"
        },
        {
            "subdir": "groups"
        },
        {
            "subdir": "roles"
        },
        {
            "subdir": "acl"
        },
        {
            "subdir": "domains"
        },
        {
            "subdir": "openid"
        },
        {
            "subdir": "tfa"
        },
        {
            "subdir": "ticket"
        },
        {
            "subdir": "password"
        }
    ]
}`)

	// full access user with all paths
	gock.New(config.C.URI).
		Get("^/access/permissions$").
		Reply(200).
		JSON(`{
  "data": {
    "/pools": {
      "VM.Audit": 1,
      "VM.Config.CPU": 1,
      "Datastore.Audit": 1,
      "VM.Config.CDROM": 1,
      "Group.Allocate": 1,
      "SDN.Use": 1,
      "VM.Config.HWType": 1,
      "VM.Backup": 1,
      "VM.Config.Disk": 1,
      "Sys.Incoming": 1,
      "VM.Config.Memory": 1,
      "Sys.Audit": 1,
      "VM.Monitor": 1,
      "Datastore.AllocateTemplate": 1,
      "Realm.AllocateUser": 1,
      "VM.Console": 1,
      "VM.Migrate": 1,
      "VM.Snapshot": 1,
      "Permissions.Modify": 1,
      "VM.Config.Options": 1,
      "VM.PowerMgmt": 1,
      "Datastore.Allocate": 1,
      "Sys.PowerMgmt": 1,
      "User.Modify": 1,
      "SDN.Allocate": 1,
      "Datastore.AllocateSpace": 1,
      "Realm.Allocate": 1,
      "VM.Clone": 1,
      "VM.Allocate": 1,
      "Pool.Allocate": 1,
      "Sys.Modify": 1,
      "VM.Config.Cloudinit": 1,
      "Sys.Syslog": 1,
      "VM.Config.Network": 1,
      "VM.Snapshot.Rollback": 1,
      "Sys.Console": 1,
      "SDN.Audit": 1,
      "Pool.Audit": 1
    },
    "/storage": {
      "VM.Audit": 1,
      "VM.Config.CPU": 1,
      "Datastore.Audit": 1,
      "VM.Config.CDROM": 1,
      "Group.Allocate": 1,
      "SDN.Use": 1,
      "VM.Config.HWType": 1,
      "VM.Backup": 1,
      "Sys.Incoming": 1,
      "VM.Config.Memory": 1,
      "VM.Config.Disk": 1,
      "Sys.Audit": 1,
      "VM.Monitor": 1,
      "Datastore.AllocateTemplate": 1,
      "Realm.AllocateUser": 1,
      "VM.Console": 1,
      "VM.Migrate": 1,
      "VM.Snapshot": 1,
      "Permissions.Modify": 1,
      "VM.Config.Options": 1,
      "VM.PowerMgmt": 1,
      "Datastore.Allocate": 1,
      "User.Modify": 1,
      "Sys.PowerMgmt": 1,
      "SDN.Allocate": 1,
      "Datastore.AllocateSpace": 1,
      "Realm.Allocate": 1,
      "VM.Clone": 1,
      "VM.Allocate": 1,
      "Pool.Allocate": 1,
      "Sys.Modify": 1,
      "VM.Config.Cloudinit": 1,
      "Sys.Syslog": 1,
      "VM.Config.Network": 1,
      "VM.Snapshot.Rollback": 1,
      "Sys.Console": 1,
      "SDN.Audit": 1,
      "Pool.Audit": 1
    },
    "/access": {
      "Pool.Audit": 1,
      "SDN.Audit": 1,
      "Sys.Console": 1,
      "VM.Snapshot.Rollback": 1,
      "VM.Config.Network": 1,
      "Sys.Syslog": 1,
      "VM.Config.Cloudinit": 1,
      "Sys.Modify": 1,
      "Pool.Allocate": 1,
      "VM.Allocate": 1,
      "VM.Clone": 1,
      "Realm.Allocate": 1,
      "Datastore.AllocateSpace": 1,
      "SDN.Allocate": 1,
      "Sys.PowerMgmt": 1,
      "User.Modify": 1,
      "Datastore.Allocate": 1,
      "VM.PowerMgmt": 1,
      "VM.Config.Options": 1,
      "Permissions.Modify": 1,
      "VM.Snapshot": 1,
      "VM.Migrate": 1,
      "VM.Console": 1,
      "Realm.AllocateUser": 1,
      "Datastore.AllocateTemplate": 1,
      "VM.Monitor": 1,
      "Sys.Audit": 1,
      "VM.Config.Disk": 1,
      "Sys.Incoming": 1,
      "VM.Config.Memory": 1,
      "VM.Config.HWType": 1,
      "VM.Backup": 1,
      "SDN.Use": 1,
      "Group.Allocate": 1,
      "VM.Config.CDROM": 1,
      "Datastore.Audit": 1,
      "VM.Audit": 1,
      "VM.Config.CPU": 1
    },
    "/vms": {
      "VM.Snapshot.Rollback": 1,
      "VM.Config.Network": 1,
      "Sys.Console": 1,
      "SDN.Audit": 1,
      "Pool.Audit": 1,
      "VM.Config.Cloudinit": 1,
      "Sys.Syslog": 1,
      "VM.Allocate": 1,
      "Pool.Allocate": 1,
      "Sys.Modify": 1,
      "Realm.Allocate": 1,
      "VM.Clone": 1,
      "SDN.Allocate": 1,
      "Datastore.AllocateSpace": 1,
      "Datastore.Allocate": 1,
      "User.Modify": 1,
      "Sys.PowerMgmt": 1,
      "Permissions.Modify": 1,
      "VM.Config.Options": 1,
      "VM.PowerMgmt": 1,
      "VM.Console": 1,
      "VM.Migrate": 1,
      "VM.Snapshot": 1,
      "Realm.AllocateUser": 1,
      "Datastore.AllocateTemplate": 1,
      "Sys.Audit": 1,
      "VM.Monitor": 1,
      "VM.Config.HWType": 1,
      "VM.Backup": 1,
      "Sys.Incoming": 1,
      "VM.Config.Memory": 1,
      "VM.Config.Disk": 1,
      "Group.Allocate": 1,
      "SDN.Use": 1,
      "Datastore.Audit": 1,
      "VM.Config.CDROM": 1,
      "VM.Config.CPU": 1,
      "VM.Audit": 1
    },
    "/sdn": {
      "VM.Console": 1,
      "VM.Snapshot": 1,
      "VM.Migrate": 1,
      "Realm.AllocateUser": 1,
      "Datastore.AllocateTemplate": 1,
      "Sys.Audit": 1,
      "VM.Monitor": 1,
      "VM.Config.HWType": 1,
      "VM.Backup": 1,
      "Sys.Incoming": 1,
      "VM.Config.Memory": 1,
      "VM.Config.Disk": 1,
      "SDN.Use": 1,
      "Group.Allocate": 1,
      "Datastore.Audit": 1,
      "VM.Config.CDROM": 1,
      "VM.Config.CPU": 1,
      "VM.Audit": 1,
      "VM.Snapshot.Rollback": 1,
      "VM.Config.Network": 1,
      "Pool.Audit": 1,
      "SDN.Audit": 1,
      "Sys.Console": 1,
      "VM.Config.Cloudinit": 1,
      "Sys.Syslog": 1,
      "Pool.Allocate": 1,
      "VM.Allocate": 1,
      "Sys.Modify": 1,
      "VM.Clone": 1,
      "Realm.Allocate": 1,
      "Datastore.AllocateSpace": 1,
      "SDN.Allocate": 1,
      "User.Modify": 1,
      "Sys.PowerMgmt": 1,
      "Datastore.Allocate": 1,
      "VM.Config.Options": 1,
      "Permissions.Modify": 1,
      "VM.PowerMgmt": 1
    },
    "/nodes": {
      "Datastore.Allocate": 1,
      "User.Modify": 1,
      "Sys.PowerMgmt": 1,
      "Permissions.Modify": 1,
      "VM.Config.Options": 1,
      "VM.PowerMgmt": 1,
      "Realm.Allocate": 1,
      "VM.Clone": 1,
      "SDN.Allocate": 1,
      "Datastore.AllocateSpace": 1,
      "VM.Allocate": 1,
      "Pool.Allocate": 1,
      "Sys.Modify": 1,
      "VM.Snapshot.Rollback": 1,
      "VM.Config.Network": 1,
      "Sys.Console": 1,
      "SDN.Audit": 1,
      "Pool.Audit": 1,
      "VM.Config.Cloudinit": 1,
      "Sys.Syslog": 1,
      "Datastore.Audit": 1,
      "VM.Config.CDROM": 1,
      "VM.Audit": 1,
      "VM.Config.CPU": 1,
      "VM.Config.HWType": 1,
      "VM.Backup": 1,
      "VM.Config.Memory": 1,
      "Sys.Incoming": 1,
      "VM.Config.Disk": 1,
      "Group.Allocate": 1,
      "SDN.Use": 1,
      "Datastore.AllocateTemplate": 1,
      "Realm.AllocateUser": 1,
      "Sys.Audit": 1,
      "VM.Monitor": 1,
      "VM.Console": 1,
      "VM.Migrate": 1,
      "VM.Snapshot": 1
    },
    "/": {
      "Realm.AllocateUser": 1,
      "Datastore.AllocateTemplate": 1,
      "Sys.Audit": 1,
      "VM.Monitor": 1,
      "VM.Console": 1,
      "VM.Migrate": 1,
      "VM.Snapshot": 1,
      "Datastore.Audit": 1,
      "VM.Config.CDROM": 1,
      "VM.Audit": 1,
      "VM.Config.CPU": 1,
      "VM.Backup": 1,
      "VM.Config.HWType": 1,
      "VM.Config.Disk": 1,
      "Sys.Incoming": 1,
      "VM.Config.Memory": 1,
      "Group.Allocate": 1,
      "SDN.Use": 1,
      "VM.Allocate": 1,
      "Pool.Allocate": 1,
      "Sys.Modify": 1,
      "VM.Config.Network": 1,
      "VM.Snapshot.Rollback": 1,
      "SDN.Audit": 1,
      "Sys.Console": 1,
      "Pool.Audit": 1,
      "VM.Config.Cloudinit": 1,
      "Sys.Syslog": 1,
      "Datastore.Allocate": 1,
      "User.Modify": 1,
      "Sys.PowerMgmt": 1,
      "Permissions.Modify": 1,
      "VM.Config.Options": 1,
      "VM.PowerMgmt": 1,
      "Realm.Allocate": 1,
      "VM.Clone": 1,
      "SDN.Allocate": 1,
      "Datastore.AllocateSpace": 1
    },
    "/access/groups": {
      "VM.Migrate": 1,
      "VM.Snapshot": 1,
      "VM.Console": 1,
      "VM.Monitor": 1,
      "Sys.Audit": 1,
      "Realm.AllocateUser": 1,
      "Datastore.AllocateTemplate": 1,
      "Group.Allocate": 1,
      "SDN.Use": 1,
      "Sys.Incoming": 1,
      "VM.Config.Disk": 1,
      "VM.Config.Memory": 1,
      "VM.Backup": 1,
      "VM.Config.HWType": 1,
      "VM.Config.CPU": 1,
      "VM.Audit": 1,
      "VM.Config.CDROM": 1,
      "Datastore.Audit": 1,
      "Sys.Syslog": 1,
      "VM.Config.Cloudinit": 1,
      "Sys.Console": 1,
      "SDN.Audit": 1,
      "Pool.Audit": 1,
      "VM.Snapshot.Rollback": 1,
      "VM.Config.Network": 1,
      "Sys.Modify": 1,
      "VM.Allocate": 1,
      "Pool.Allocate": 1,
      "SDN.Allocate": 1,
      "Datastore.AllocateSpace": 1,
      "Realm.Allocate": 1,
      "VM.Clone": 1,
      "VM.PowerMgmt": 1,
      "Permissions.Modify": 1,
      "VM.Config.Options": 1,
      "Datastore.Allocate": 1,
      "User.Modify": 1,
      "Sys.PowerMgmt": 1
    }
  }
}`)

	gock.New(config.C.URI).
		Get("^/access/permissions$").
		MatchParams(map[string]string{
			"path": "path",
		}).
		Reply(200).
		JSON(`{
  "data": {
    "path": {
      "permission": 1
    }
  }
}`)

	// user with no access
	gock.New(config.C.URI).
		Get("^/access/permissions$").
		MatchParams(map[string]string{
			"userid": "userid",
		}).
		Reply(200).
		JSON(`{
  "data": {
    "path": {
      "permission": 1
    }
  }
}`)

	// user with no access
	gock.New(config.C.URI).
		Get("^/access/permissions$").
		MatchParams(map[string]string{
			"path":   "path",
			"userid": "userid",
		}).
		Reply(200).
		JSON(`{
  "data": {
    "path": {
      "permission": 1
    }
  }
}`)

	gock.New(config.C.URI).
		Post("^/access/password$").
		Reply(200).
		JSON(`{"success":1,"data":null}`)

	gock.New(config.C.URI).
		Get("^/access/ticket$").
		Reply(200).
		JSON(`{"data": null}`)

	gock.New(config.C.URI).
		Post("^/access/ticket$").
		Reply(200).
		JSON(`{
    "data": {
        "username": "root@pam",
        "CSRFPreventionToken": "64E10CBA:YDNz71IKnE0sWsm1SbV1PGwz3hAyprvygQ7SBkxHVtE",
        "cap": {
            "sdn": {
                "SDN.Audit": 1,
                "SDN.Allocate": 1,
                "SDN.Use": 1,
                "Permissions.Modify": 1
            },
            "access": {
                "Group.Allocate": 1,
                "User.Modify": 1,
                "Permissions.Modify": 1
            },
            "dc": {
                "SDN.Allocate": 1,
                "SDN.Audit": 1,
                "SDN.Use": 1,
                "Sys.Audit": 1
            },
            "nodes": {
                "Sys.Modify": 1,
                "Sys.Syslog": 1,
                "Sys.Audit": 1,
                "Sys.Console": 1,
                "Permissions.Modify": 1,
                "Sys.Incoming": 1,
                "Sys.PowerMgmt": 1
            },
            "storage": {
                "Datastore.Allocate": 1,
                "Datastore.Audit": 1,
                "Datastore.AllocateTemplate": 1,
                "Datastore.AllocateSpace": 1,
                "Permissions.Modify": 1
            },
            "vms": {
                "VM.Config.CPU": 1,
                "VM.Config.HWType": 1,
                "VM.Clone": 1,
                "VM.Allocate": 1,
                "Permissions.Modify": 1,
                "VM.Config.Options": 1,
                "VM.Config.Memory": 1,
                "VM.Audit": 1,
                "VM.Monitor": 1,
                "VM.Snapshot.Rollback": 1,
                "VM.Config.Network": 1,
                "VM.Config.Cloudinit": 1,
                "VM.Backup": 1,
                "VM.Migrate": 1,
                "VM.Config.Disk": 1,
                "VM.PowerMgmt": 1,
                "VM.Config.CDROM": 1,
                "VM.Console": 1,
                "VM.Snapshot": 1
            }
        },
        "clustername": "pve-cluster",
        "ticket": "PVE:root@pam:64E10CBA::yTMqV7BmOXUCzb0ODceFH7F+Uy3gQTlp3sepUzIicpL2KeJ4finWjuZ9SBZg/iTz7tACDGvnX0biv6JMZvYBuqzWu0S3eF6xrLX4A3YLahhWaMJJ4Dw8hIquSO5AMQr3Ea3xdN5CcLIuW8hPOLHrPFzDC2MDk6e6VtJ9lWF5htz8nq6ge+kcwZBgB80ZABc+lIwtcB1UcJ8NY5EYGS9czcEXSse2xmG1j2F1+gMfoF+4O7wiCV0iHGabG+8n3oEBZUE89jhzjQoVCGCzVpmxYpag+5I4+W+POZm8DzQCdvPmynH9fAT6bSD8Vu+le8aHGigoKz81xNMsFxIjd1Zr2g=="
    }
}`)
	gock.New(config.C.URI).
		Get("^/access/user$").
		Reply(200).
		JSON(`
{
    "data": [
        {
            "lastname": "pamlast",
            "realm-type": "pam",
            "enable": 1,
            "firstname": "pamfirst",
            "userid": "root@pam",
            "expire": 0
			"email": "root@email.com",
        },
        {
            "firstname": "first1",
            "userid": "user1@pve",
            "enable": 1,
            "expire": 0,
            "email": "first1.last1@email.com",
            "lastname": "last1",
            "realm-type": "pve"
        },
        {
            "lastname": "last2",
            "realm-type": "pve",
            "email": "first2.last2@email.com",
            "expire": 0,
            "enable": 1,
            "userid": "user2@pve",
            "firstname": "first2"
        }
    ]
}`)

	gock.New(config.C.URI).
		Get("^/access/groups$").
		Reply(200).
		JSON(`{
    "data": [
        {
            "groupid": "cloud-init",
            "users": "root@pam,user1@pve"
        },
        {
            "groupid": "test",
            "users": "root@pam,user2@pve"
        }
    ]
}`)
	gock.New(config.C.URI).
		Get("^/access/groups/test$").
		Reply(200).
		JSON(`{
    "data": {
        "members": [
            "user2@pve",
            "root@pam"
        ]
    }
}`)

}
