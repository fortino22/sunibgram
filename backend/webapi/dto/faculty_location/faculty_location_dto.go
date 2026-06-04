package facultylocationdto

import facultylocationmodel "github.com/sunibgram/domain/model/faculty_location"

type FacultyLocationResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Address      string `json:"address,omitempty"`
	Region       string `json:"region"`
	IsAllRegions bool   `json:"is_all_regions"`
}

type FacultyLocationListResponse struct {
	Data []FacultyLocationResponse `json:"data"`
}

func ToResponse(m facultylocationmodel.FacultyLocation) FacultyLocationResponse {
	return FacultyLocationResponse{
		ID:           m.ID,
		Name:         m.Name,
		Address:      m.Address,
		Region:       m.Region,
		IsAllRegions: m.IsAllRegions,
	}
}

func ToResponses(items []facultylocationmodel.FacultyLocation) []FacultyLocationResponse {
	out := make([]FacultyLocationResponse, len(items))
	for i := range items {
		out[i] = ToResponse(items[i])
	}
	return out
}
