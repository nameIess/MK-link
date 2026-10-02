package core

type LinkType string

const (
    LinkTypeSymbolic LinkType = "symbolic"
    LinkTypeHard LinkType = "hard"
    LinkTypeJunction LinkType = "junction"
)

type AppInfo struct {
    Name string
    Version string
    Description string
}

type PathInfo struct {
    Path string
    Exists bool
    IsFile bool
    IsDir bool
    IsReparse bool
    Size int64
}

type LinkTypeInfo struct {
    ID LinkType
    Name string
    Summary string
    BestFor string
    TargetKind string
    Description string
}

type CreateLinkRequest struct {
    TargetPath string
    DestinationDir string
    LinkName string
    LinkType LinkType
}

type ValidatedCreateRequest struct {
    TargetPath string
    LinkPath string
    LinkType LinkType
}

type ValidationResult struct {
    Valid bool
    Message string
    LinkPath string
    Conflict bool
}

type CreateLinkResult struct {
    Success bool
    LinkPath string
    Error string
    NeedsElevation bool
}

type HealthStatus struct{ Ready bool }
