package ccv3

import (
	"bytes"
	"encoding/json"

	"code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/internal"
)

// RelationshipList represents a one to many relationship.
type RelationshipList struct {
	GUIDs []string
}

func (r RelationshipList) MarshalJSON() ([]byte, error) {
	var ccRelationship struct {
		Data []map[string]string `json:"data"`
	}

	for _, guid := range r.GUIDs {
		ccRelationship.Data = append(
			ccRelationship.Data,
			map[string]string{
				"guid": guid,
			})
	}

	return json.Marshal(ccRelationship)
}

func (r *RelationshipList) UnmarshalJSON(data []byte) error {
	var ccRelationship struct {
		Data []map[string]string `json:"data"`
	}

	err := json.Unmarshal(data, &ccRelationship)
	if err != nil {
		return err
	}

	for _, partner := range ccRelationship.Data {
		r.GUIDs = append(r.GUIDs, partner["guid"])
	}
	return nil
}

// EntitleIsolationSegmentToOrganizations will create a link between the
// isolation segment and the list of organizations provided.
func (client *Client) EntitleIsolationSegmentToOrganizations(isolationSegmentGUID string, organizationGUIDs []string) (RelationshipList, Warnings, error) {
	body, err := json.Marshal(RelationshipList{GUIDs: organizationGUIDs})
	if err != nil {
		return RelationshipList{}, nil, err
	}

	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.PostIsolationSegmentRelationshipOrganizationsRequest,
		URIParams:   internal.Params{"guid": isolationSegmentGUID},
		Body:        bytes.NewBuffer(body),
	})

	var relationships RelationshipList
	response := cloudcontroller.Response{
		Result: &relationships,
	}

	err = client.connection.Make(request, &response)
	return relationships, response.Warnings, err
}

// RevokeIsolationSegmentFromOrganization will delete the relationship between
// the isolation segment and the organization provided.
func (client *Client) RevokeIsolationSegmentFromOrganization(isolationSegmentGUID string, organizationGUID string) (Warnings, error) {
	request, err := client.newHTTPRequest(requestOptions{
		RequestName: internal.DeleteIsolationSegmentRelationshipOrganizationRequest,
		URIParams:   internal.Params{"guid": isolationSegmentGUID, "org_guid": organizationGUID},
	})
	if err != nil {
		return nil, err
	}

	var response cloudcontroller.Response
	err = client.connection.Make(request, &response)

	return response.Warnings, err
}
