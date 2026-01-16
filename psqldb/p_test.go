package psqldb

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestGetIDFromInData(t *testing.T) {
	var op Operation
	op.InData = json.RawMessage(`{"id": "abcd:12345:6789:uri", "other": "whatever it is ther"}`)
	op.CMStatus = 200
	op.LDStatus = 201
	op.Method = "POST"
	op.RequestUri = "/ngsi-ld/v1/entities"

	eid := getEntityID(&op)
	if eid != "abcd:12345:6789:uri" {
		t.Errorf("expected ID is %s instead of the expected one", eid)
	}

	fmt.Println(eid)
}

func TestGEtIDFromURI(t *testing.T) {
	exptd_id := "urn:abcd:12345:6789:emf1"
	var op Operation
	op.InData = json.RawMessage(`{"other": "some value"}`)
	op.RequestUri = "/ngsi-ld/v1/entities/" + exptd_id
	op.CMStatus = 200
	op.LDStatus = 201
	op.Method = "POST"

	if nid := getEntityID(&op); nid != exptd_id {
		t.Errorf("the ID is %s instead of the expected one %s", nid, exptd_id)
	}

	op.RequestUri = "/ngsi-ld/v1/entities/" + exptd_id + "/attrs"
	if nid := getEntityID(&op); nid != exptd_id {
		t.Errorf("the ID is %s instead of the expected one %s", nid, exptd_id)
	}

	op.RequestUri = "/ngsi-ld/v1/entities/" + exptd_id + "/attrs/"
	if nid := getEntityID(&op); nid != exptd_id {
		t.Errorf("the ID is %s instead of the expected one %s", nid, exptd_id)
	}
}
