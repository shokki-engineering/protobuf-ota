// Copyright 2026 Shokki Engineering.
// SPDX-License-Identifier: Apache-2.0

package model

// Resource is one row of the resources table: an OTA type that becomes an
// AIP resource (CLAUDE.md rule 7), where it sits, and what of it belongs to
// other resources. docs/resources.md is this table rendered for a reader.
type Resource struct {
	Spec   string // "hotel" or "common"
	Domain string // "" for a root or a common package
	Name   string // the message, e.g. "RatePlan"

	// Root is the element path of the OTA type the resource carries, from
	// the message that writes it.
	Root string

	// Identity is the root's attribute that is the resource id, the last
	// segment of its name; empty when the server assigns the id.
	Identity string

	Parent    string // the parent resource's Name; "" for a root collection
	Singleton bool

	// Carve lists the root's members that are not fields of this resource:
	// child resources, and references to the parent its name already
	// carries. Element names, or attribute names written "@Name".
	Carve []string

	// Merge is a second OTA type the resource also carries (docs/resources.md,
	// RoomType): its members join the first's, less its identity and carves.
	Merge *Merge

	// Renames are field renames that apply on this resource only, keyed by
	// OTA path -- HotelName is a hotel's display_name, but a reference to a
	// hotel elsewhere carries it as hotel_display_name.
	Renames map[string]string
}

// Merge is a second source for one resource.
type Merge struct {
	Root     string   // element path, like Resource.Root
	Identity string   // its attribute naming the same id as the first's
	Carve    []string // as Resource.Carve
	// Drop lists members the first source already carries, by OTA path. A
	// field name both sources produce is a build error until one is dropped
	// here -- the merges table CLAUDE.md rule 2 names.
	Drop []string
}

// resources is the table, assembled from one file per area so each stays
// readable. Order is irrelevant; parents are found by name.
var resources = concat(propertyResources, revenueResources, bookingResources, commonResources)

func concat(lists ...[]*Resource) []*Resource {
	var out []*Resource
	for _, l := range lists {
		out = append(out, l...)
	}
	return out
}

// hdc is the hotel descriptive content element every property resource is
// carved from.
const hdc = "OTA_HotelDescriptiveContentNotifRQ/HotelDescriptiveContents/HotelDescriptiveContent"

var propertyResources = []*Resource{
	{
		Spec: "hotel", Name: "Hotel", Root: hdc, Identity: "HotelCode",
		Carve: []string{
			"FacilityInfo", "Policies", "Promotions", "MultimediaDescriptions",
			"ContactInfos", "AreaInfo", "AffiliationInfo", "EnvironmentalImpact",
		},
		Renames: map[string]string{"HotelReferenceGroup/@HotelName": "display_name"},
	},
	{
		Spec: "hotel", Domain: "property", Name: "RoomType", Parent: "Hotel",
		Root: hdc + "/FacilityInfo/GuestRooms/GuestRoom", Identity: "Code",
		// OTA describes a room type in descriptive content and sells it as a
		// SellableProduct under the code every rate and count cites. One
		// resource, so one room type has one name.
		Merge: &Merge{
			Root:     "OTA_HotelInvNotifRQ/SellableProducts/SellableProduct",
			Identity: "InvTypeCode",
			Carve:    []string{"MeetingRooms", "InventoryBlock", "@IsRoom"},
		},
		Renames: map[string]string{"FacilityInfoType/GuestRooms/GuestRoom/@RoomTypeName": "display_name"},
	},
	{
		Spec: "hotel", Domain: "property", Name: "MeetingRoom", Parent: "Hotel",
		Root: hdc + "/FacilityInfo/MeetingRooms/MeetingRoom", Identity: "ID",
	},
	{
		Spec: "hotel", Domain: "property", Name: "Restaurant", Parent: "Hotel",
		Root: hdc + "/FacilityInfo/Restaurants/Restaurant", Identity: "ID",
		Renames: map[string]string{"RestaurantType/@RestaurantName": "display_name"},
	},
	{Spec: "hotel", Domain: "property", Name: "Policy", Parent: "Hotel", Root: hdc + "/Policies/Policy"},
	{
		Spec: "hotel", Domain: "property", Name: "Promotion", Parent: "Hotel",
		Root: hdc + "/Promotions/Promotion", Identity: "PromotionCode",
	},
	{
		Spec: "hotel", Domain: "property", Name: "Multimedia", Parent: "Hotel",
		Root: hdc + "/MultimediaDescriptions/MultimediaDescription",
	},
	{
		Spec: "hotel", Domain: "property", Name: "Contact", Parent: "Hotel",
		Root: hdc + "/ContactInfos/ContactInfo", Identity: "ContactProfileID",
	},
	{Spec: "hotel", Domain: "property", Name: "AreaInfo", Parent: "Hotel", Root: hdc + "/AreaInfo", Singleton: true},
	{Spec: "hotel", Domain: "property", Name: "AffiliationInfo", Parent: "Hotel", Root: hdc + "/AffiliationInfo", Singleton: true},
	{
		Spec: "hotel", Domain: "property", Name: "EnvironmentalImpact", Parent: "Hotel",
		Root: hdc + "/EnvironmentalImpact", Singleton: true,
	},
}
