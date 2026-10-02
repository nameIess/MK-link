export namespace core {
	
	export class AppInfo {
	    Name: string;
	    Version: string;
	    Description: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Version = source["Version"];
	        this.Description = source["Description"];
	    }
	}
	export class CreateLinkRequest {
	    TargetPath: string;
	    DestinationDir: string;
	    LinkName: string;
	    LinkType: string;
	
	    static createFrom(source: any = {}) {
	        return new CreateLinkRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.TargetPath = source["TargetPath"];
	        this.DestinationDir = source["DestinationDir"];
	        this.LinkName = source["LinkName"];
	        this.LinkType = source["LinkType"];
	    }
	}
	export class CreateLinkResult {
	    Success: boolean;
	    LinkPath: string;
	    Error: string;
	    NeedsElevation: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CreateLinkResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Success = source["Success"];
	        this.LinkPath = source["LinkPath"];
	        this.Error = source["Error"];
	        this.NeedsElevation = source["NeedsElevation"];
	    }
	}
	export class HealthStatus {
	    Ready: boolean;
	
	    static createFrom(source: any = {}) {
	        return new HealthStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Ready = source["Ready"];
	    }
	}
	export class LinkTypeInfo {
	    ID: string;
	    Name: string;
	    Summary: string;
	    BestFor: string;
	    TargetKind: string;
	    Description: string;
	
	    static createFrom(source: any = {}) {
	        return new LinkTypeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Name = source["Name"];
	        this.Summary = source["Summary"];
	        this.BestFor = source["BestFor"];
	        this.TargetKind = source["TargetKind"];
	        this.Description = source["Description"];
	    }
	}
	export class PathInfo {
	    Path: string;
	    Exists: boolean;
	    IsFile: boolean;
	    IsDir: boolean;
	    IsReparse: boolean;
	    Size: number;
	
	    static createFrom(source: any = {}) {
	        return new PathInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Path = source["Path"];
	        this.Exists = source["Exists"];
	        this.IsFile = source["IsFile"];
	        this.IsDir = source["IsDir"];
	        this.IsReparse = source["IsReparse"];
	        this.Size = source["Size"];
	    }
	}
	export class ValidationResult {
	    Valid: boolean;
	    Message: string;
	    LinkPath: string;
	    Conflict: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Valid = source["Valid"];
	        this.Message = source["Message"];
	        this.LinkPath = source["LinkPath"];
	        this.Conflict = source["Conflict"];
	    }
	}

}

