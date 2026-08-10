package sightengine

type CheckResponse struct {
	Status string `json:"status"`
	Type   struct {
		AIGenerated  float64            `json:"ai_generated"`
		AIGenerators map[string]float64 `json:"ai_generators"`
	} `json:"type"`
}
