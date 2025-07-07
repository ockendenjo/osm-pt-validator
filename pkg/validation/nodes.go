package validation

import (
	"context"
	"fmt"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func (v *Validator) validateRelationNodes(ctx context.Context, re osm.Relation) ([]ValidationError, error) {
	nodeIds := []int64{}
	nodes := []osm.Member{}
	validationErrors := []ValidationError{}

	for _, member := range re.Members {
		if member.Type == "node" {
			nodeIds = append(nodeIds, member.Ref)
			nodes = append(nodes, member)
		}
	}

	nodesMap := v.osmClient.LoadNodes(ctx, nodeIds)

	for k, way := range nodesMap {
		if way == nil {
			return nil, fmt.Errorf("failed to load node %d", k)
		}
	}

	for _, node := range nodes {
		nodeObj := nodesMap[node.Ref]
		if node.RoleIsPlatform() {
			validationErrors = append(validationErrors, validatePlatformNode(nodeObj, v.config.NaptanPlatformTags)...)
		}

		if node.RoleIsStop() {
			validationErrors = append(validationErrors, validateStopNode(nodeObj)...)
		}
	}

	return validationErrors, nil
}

func validatePlatformNode(node *osm.Node, checkNaptan bool) []ValidationError {
	validationErrors := []ValidationError{}

	pt, found := node.Tags["public_transport"]
	if !found {
		validationErrors = append(validationErrors, ValidationError{URL: node.GetElementURL(), Message: "node is missing public_transport tag"})
	} else if pt != "platform" {
		validationErrors = append(validationErrors, ValidationError{URL: node.GetElementURL(), Message: "node should have public_transport=platform"})
	}

	_, found = node.Tags["disused:highway"]
	if found {
		validationErrors = append(validationErrors, ValidationError{URL: node.GetElementURL(), Message: "node has disused:highway tag"})
	}

	//Don't require the highway tag to be present - Naptan imported stops don't have it set (to prevent rendering)
	highway, found := node.Tags["highway"]
	if found && highway != "bus_stop" {
		validationErrors = append(validationErrors, ValidationError{URL: node.GetElementURL(), Message: "node should have highway=bus_stop"})
	}

	_, found = node.Tags["name"]
	if !found {
		validationErrors = append(validationErrors, ValidationError{URL: node.GetElementURL(), Message: "node is missing name tag"})
	}

	if checkNaptan {
		missingTagErrors := checkTagsPresent(node, "naptan:AtcoCode")
		validationErrors = append(validationErrors, missingTagErrors...)
	}

	return validationErrors
}

func validateStopNode(node *osm.Node) []ValidationError {
	validationErrors := []ValidationError{}

	pt, found := node.Tags["public_transport"]
	if !found {
		validationErrors = append(validationErrors, ValidationError{URL: node.GetElementURL(), Message: "node is missing public_transport tag"})
	} else if pt != "stop_position" {
		validationErrors = append(validationErrors, ValidationError{URL: node.GetElementURL(), Message: "node should have public_transport=stop_position"})
	}

	bus, found := node.Tags["bus"]
	if found && bus != "yes" {
		validationErrors = append(validationErrors, ValidationError{URL: node.GetElementURL(), Message: "node should have bus=yes"})
	}

	name, found := node.Tags["name"]
	if !found || len(name) < 1 {
		//TODO: Add better validation to check if stop is part of public_transport=stop_area
		//validationErrors = append(validationErrors, ValidationError{URL: node.GetElementURL(), Message: "node is missing name tag"})
	}

	return validationErrors
}
