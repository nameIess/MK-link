import {
  ChooseDirectory,
  ChooseFile,
  CreateLink,
  GetLinkTypes,
  OpenInExplorer,
  ValidateDestination,
  ValidateTarget,
} from "../wailsjs/go/main/App";
import type { CreateLinkRequest, CreateLinkResult, LinkType, LinkTypeInfo, PathInfo, ValidationResult } from "./types";

export const chooseTargetFile = ChooseFile;
export const chooseTargetDirectory = () => ChooseDirectory("Choose target folder");
export const chooseDestination = () => ChooseDirectory("Choose destination folder");
export const inspectTarget = (path: string): Promise<PathInfo> => ValidateTarget(path);
export const getTypes = async (path: string): Promise<LinkTypeInfo[]> => {
  const types = await GetLinkTypes(path);

  return types.map((type) => ({
    ...type,
    ID: type.ID as LinkType,
  }));
};
export const validateDestination = (dir: string, name: string, type: LinkType): Promise<ValidationResult> =>
  ValidateDestination(dir, name, type);
export const createLink = (req: CreateLinkRequest): Promise<CreateLinkResult> => CreateLink(req);
export const openInExplorer = OpenInExplorer;
