package airtable

import (
	"fmt"
	"os"
	"testing"

	. "github.com/Antfood/airgo/testutils/testutils"
)

type testCreatedBy struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type testRecordSchema struct {
	Title       string        `json:"rec.Recording Title"`
	CyRecording []string      `json:"cy.Recording"`
	CreatedBy   testCreatedBy `json:"at.CreatedBy"`
}

func TestGet(t *testing.T) {
	json, err := os.ReadFile("./testdata/get_recording.json")
	Ok(t, err)

	client := newMockClient(200, json, nil)
	Configure(client, "mock_token")

   table := NewTable[testRecordSchema]("Base_id", "Table_id")
	r, err := table.Get("recording_id")

	Assert(t, err == nil, "Should not return an error: %s", err)
	Assert(t, r.Fields.Title == "Soliloquy (Original)", "Should have same title")
	Assert(t, r.Fields.CyRecording[0] == "recBkifuU3RrwzFTI", "Should have same cy.Recording Id")

	Equals(t, r.Fields.CreatedBy,
		testCreatedBy{Id: "usrUluXd4j2EEgZnt",
			Email: "pedro@antfood.com",
			Name:  "Pedro Botsaris"})

	client = newMockClient(404, json, fmt.Errorf("Not Found"))
	Configure(client, "mock_token")

	_, err = table.Get("recording_id")

	err_msg := "airtable.Get: Not Found"
	Assert(t, err != nil, "Should return error")
	Assert(t, err.Error() == err_msg, "Error was '%s' should be '%s'", err.Error(), err_msg)
}
