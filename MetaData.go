package main


// Metadata
type MetadataEntry struct {
	Attrs MetadataEntryAttribute `json:"attrs"`
	Stacks []Stack `json:"stacks"`
}

type MetadataEntryAttribute struct {
	Version int `json:"version"`
	NumStacks int `json:"num_stacks"`
}

type Stack struct {
	Levels []Level `json:"levels"`
	Attrs  StackAttrs `json:"attrs"`
}

type StackAttrs struct {
	Description string `json:"description"`
	NumLevels int `json:"num_levels"`
	NumSlices int `json:"num_slices"`
}

type Level struct {
	Attrs LevelAttrs `json:"attrs"`
}

type LevelAttrs struct {
	NumSlices int `json:"num_slices"`
	NumXTiles int `json:"num_x_tiles"`
	NumYTiles int `json:"num_y_tiles"`
}
