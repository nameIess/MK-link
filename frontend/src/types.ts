export type LinkType = "symbolic" | "hard" | "junction";

export interface PathInfo {
  Path: string;
  Exists: boolean;
  IsFile: boolean;
  IsDir: boolean;
  IsReparse: boolean;
  Size: number;
}
export interface LinkTypeInfo {
  ID: LinkType;
  Name: string;
  Summary: string;
  BestFor: string;
  TargetKind: string;
  Description: string;
}
export interface CreateLinkRequest {
  TargetPath: string;
  DestinationDir: string;
  LinkName: string;
  LinkType: LinkType;
}
export interface CreateLinkResult {
  Success: boolean;
  LinkPath: string;
  Error: string;
  NeedsElevation: boolean;
}
export interface ValidationResult {
  Valid: boolean;
  Message: string;
  LinkPath: string;
  Conflict: boolean;
}
