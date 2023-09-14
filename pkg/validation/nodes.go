package validation

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/ockendenjo/osm-pt-validator/pkg/osm"
)

func (v *Validator) validateRelationNodes(ctx context.Context, re osm.RelationElement) ([]string, error) {
	nodeIds := []int64{}
	nodes := []osm.Member{}
	validationErrors := []string{}

	for _, member := range re.Members {
		if member.Type == "node" {
			nodeIds = append(nodeIds, member.Ref)
			nodes = append(nodes, member)
		}
	}

	nodesMap := v.osmClient.LoadNodes(ctx, nodeIds)

	for k, way := range nodesMap {
		if way == nil {
			return []string{}, fmt.Errorf("failed to load node %d", k)
		}
	}

	for _, node := range nodes {
		role := node.Role
		nodeObj := nodesMap[node.Ref]
		if role == "platform" || role == "platform_exit_only" || role == "platform_entry_only" {
			validationErrors = append(validationErrors, validatePlatformNode(nodeObj, v.config.NaptanPlatformTags)...)
		}

		if role == "stop" || role == "stop_exit_only" || role == "stop_entry_only" {
			validationErrors = append(validationErrors, validateStopNode(nodeObj)...)
		}
	}

	return validationErrors, nil
}

func shouldCheckNaptanTags() bool {
	//Gradual roll out
	threshold := float64(1695168000-time.Now().Unix()) / (24 * 60 * 60 * 10)
	return rand.Float64() > threshold
}

func validatePlatformNode(node *osm.Node, checkNaptan bool) []string {
	validationErrors := []string{}

	for _, element := range node.Elements {
		pt, found := element.Tags["public_transport"]
		if !found {
			validationErrors = append(validationErrors, fmt.Sprintf("node is missing public_transport tag - %s", element.GetElementURL()))
		} else if pt != "platform" {
			validationErrors = append(validationErrors, fmt.Sprintf("node should have public_transport=platform - %s", element.GetElementURL()))
		}

		_, found = element.Tags["disused:highway"]
		if found {
			validationErrors = append(validationErrors, fmt.Sprintf("node has disused:highway tag - %s", element.GetElementURL()))
		}

		//Don't require the highway tag to be present - Naptan imported stops don't have it set (to prevent rendering)
		highway, found := element.Tags["highway"]
		if found && highway != "bus_stop" {
			validationErrors = append(validationErrors, fmt.Sprintf("node should have highway=bus_stop - %s", element.GetElementURL()))
		}

		if checkNaptan {
			missingTagErrors := checkTagsPresent(element, "naptan:AtcoCode")
			validationErrors = append(validationErrors, missingTagErrors...)
		}
	}

	return validationErrors
}

func validateStopNode(node *osm.Node) []string {
	validationErrors := []string{}

	for _, element := range node.Elements {
		pt, found := element.Tags["public_transport"]
		if !found {
			validationErrors = append(validationErrors, fmt.Sprintf("node is missing public_transport tag - %s", element.GetElementURL()))
		} else if pt != "stop_position" {
			validationErrors = append(validationErrors, fmt.Sprintf("node should have public_transport=stop_position - %s", element.GetElementURL()))
		}

		bus, found := element.Tags["bus"]
		if found && bus != "yes" {
			validationErrors = append(validationErrors, fmt.Sprintf("node should have bus=yes - %s", element.GetElementURL()))
		}

		name, found := element.Tags["name"]
		if !found || len(name) < 1 {
			validationErrors = append(validationErrors, fmt.Sprintf("node is missing name tag - %s", element.GetElementURL()))
		}
	}

	return validationErrors
}
