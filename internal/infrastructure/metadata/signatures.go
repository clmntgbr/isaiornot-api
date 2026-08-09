package metadata

type SoftwareSignature struct {
	Pattern  string
	Label    string
	ScoreMin int
	ScoreMax int
}

var SoftwareSignatures = []SoftwareSignature{
	{Pattern: "midjourney", Label: "Midjourney", ScoreMin: 78, ScoreMax: 85},
	{Pattern: "stable diffusion", Label: "Stable Diffusion", ScoreMin: 75, ScoreMax: 85},
	{Pattern: "stablediffusion", Label: "Stable Diffusion", ScoreMin: 75, ScoreMax: 85},
	{Pattern: "automatic1111", Label: "Automatic1111", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "comfyui", Label: "ComfyUI", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "dall-e", Label: "DALL-E", ScoreMin: 78, ScoreMax: 85},
	{Pattern: "dall·e", Label: "DALL-E", ScoreMin: 78, ScoreMax: 85},
	{Pattern: "dalle", Label: "DALL-E", ScoreMin: 78, ScoreMax: 85},
	{Pattern: "openai", Label: "OpenAI", ScoreMin: 70, ScoreMax: 80},
	{Pattern: "adobe firefly", Label: "Adobe Firefly", ScoreMin: 75, ScoreMax: 85},
	{Pattern: "firefly", Label: "Adobe Firefly", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "leonardo.ai", Label: "Leonardo.ai", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "leonardo ai", Label: "Leonardo.ai", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "ideogram", Label: "Ideogram", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "flux", Label: "Flux", ScoreMin: 70, ScoreMax: 80},
	{Pattern: "black forest labs", Label: "Black Forest Labs", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "novelai", Label: "NovelAI", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "bing image creator", Label: "Bing Image Creator", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "copilot", Label: "Microsoft Copilot", ScoreMin: 70, ScoreMax: 80},
	{Pattern: "imagen", Label: "Google Imagen", ScoreMin: 72, ScoreMax: 82},
	{Pattern: "gemini", Label: "Google Gemini", ScoreMin: 70, ScoreMax: 80},
	{Pattern: "playground ai", Label: "Playground AI", ScoreMin: 70, ScoreMax: 80},
	{Pattern: "nightcafe", Label: "NightCafe", ScoreMin: 70, ScoreMax: 80},
	{Pattern: "craiyon", Label: "Craiyon", ScoreMin: 68, ScoreMax: 78},
}

var C2PAAIIndicators = []string{
	"trainedalgorithmicmedia",
	"compositewithtrainedalgorithmicmedia",
	"c2pa.training",
	"algorithmicmedia",
	"digitalsource",
	"trained algorithmic",
}

var C2PAGenerators = []string{
	"openai",
	"adobe firefly",
	"firefly",
	"google",
	"dall-e",
	"dalle",
	"midjourney",
	"stable diffusion",
}

var PromptIndicators = []string{
	"prompt",
	"negative_prompt",
	"negativeprompt",
	"seed",
	"cfg_scale",
	"cfgscale",
	"sampler",
	"steps",
	"model hash",
	"modelhash",
}

var NeutralSoftware = []string{
	"adobe photoshop",
	"photoshop",
	"lightroom",
	"gimp",
	"affinity photo",
	"capture one",
	"pixelmator",
	"apple photos",
	"preview",
	"canon",
	"nikon",
	"sony",
	"fujifilm",
}
