export namespace core {
	
	export class ExportOptions {
	    single: boolean;
	    layout: boolean;
	    social: boolean;
	    idphoto: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ExportOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.single = source["single"];
	        this.layout = source["layout"];
	        this.social = source["social"];
	        this.idphoto = source["idphoto"];
	    }
	}
	export class ExportResult {
	    ok: boolean;
	    dir: string;
	
	    static createFrom(source: any = {}) {
	        return new ExportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.dir = source["dir"];
	    }
	}
	export class GenerateParams {
	    template: string;
	    bgMode: string;
	    bgColor: string;
	    bgPreset: string;
	    customWidth: number;
	    customHeight: number;
	    beautyStrength: number;
	    eyeSharp: number;
	    skinBright: number;
	    enableMakeup: boolean;
	    lipStrength: number;
	    refineBrow: boolean;
	    browFill: number;
	    refineHair: boolean;
	    hairColorUniform: number;
	    hairlineRepair: number;
	    enableCloth: boolean;
	    clothType: string;
	    clothFit: number;
	    addWatermark: boolean;
	    genPrintLayout: boolean;
	    paperSize: string;
	    enableTargetFileSize: boolean;
	    targetFileSize: number;
	    maskFeather: number;
	    sourceImg: string;
	
	    static createFrom(source: any = {}) {
	        return new GenerateParams(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.template = source["template"];
	        this.bgMode = source["bgMode"];
	        this.bgColor = source["bgColor"];
	        this.bgPreset = source["bgPreset"];
	        this.customWidth = source["customWidth"];
	        this.customHeight = source["customHeight"];
	        this.beautyStrength = source["beautyStrength"];
	        this.eyeSharp = source["eyeSharp"];
	        this.skinBright = source["skinBright"];
	        this.enableMakeup = source["enableMakeup"];
	        this.lipStrength = source["lipStrength"];
	        this.refineBrow = source["refineBrow"];
	        this.browFill = source["browFill"];
	        this.refineHair = source["refineHair"];
	        this.hairColorUniform = source["hairColorUniform"];
	        this.hairlineRepair = source["hairlineRepair"];
	        this.enableCloth = source["enableCloth"];
	        this.clothType = source["clothType"];
	        this.clothFit = source["clothFit"];
	        this.addWatermark = source["addWatermark"];
	        this.genPrintLayout = source["genPrintLayout"];
	        this.paperSize = source["paperSize"];
	        this.enableTargetFileSize = source["enableTargetFileSize"];
	        this.targetFileSize = source["targetFileSize"];
	        this.maskFeather = source["maskFeather"];
	        this.sourceImg = source["sourceImg"];
	    }
	}
	export class Report {
	    faceOk: boolean;
	    faceScore: number;
	    isRephoto: boolean;
	    isAiImage: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Report(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.faceOk = source["faceOk"];
	        this.faceScore = source["faceScore"];
	        this.isRephoto = source["isRephoto"];
	        this.isAiImage = source["isAiImage"];
	    }
	}
	export class ResultBundle {
	    single: string;
	    layout: string;
	    social: string;
	    idphoto: string;
	
	    static createFrom(source: any = {}) {
	        return new ResultBundle(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.single = source["single"];
	        this.layout = source["layout"];
	        this.social = source["social"];
	        this.idphoto = source["idphoto"];
	    }
	}
	export class GenerateResult {
	    originImg: string;
	    resultImg: string;
	    results: ResultBundle;
	    faceBox: number[];
	    landmarks: number[];
	    report: Report;
	
	    static createFrom(source: any = {}) {
	        return new GenerateResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.originImg = source["originImg"];
	        this.resultImg = source["resultImg"];
	        this.results = this.convertValues(source["results"], ResultBundle);
	        this.faceBox = source["faceBox"];
	        this.landmarks = source["landmarks"];
	        this.report = this.convertValues(source["report"], Report);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LoadImageResult {
	    imgBase64: string;
	    faceBox: number[];
	    landmarks: number[];
	    report: Report;
	
	    static createFrom(source: any = {}) {
	        return new LoadImageResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.imgBase64 = source["imgBase64"];
	        this.faceBox = source["faceBox"];
	        this.landmarks = source["landmarks"];
	        this.report = this.convertValues(source["report"], Report);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PaperSpec {
	    value: string;
	    title: string;
	    desc: string;
	    widthMm?: number;
	    heightMm?: number;
	
	    static createFrom(source: any = {}) {
	        return new PaperSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.title = source["title"];
	        this.desc = source["desc"];
	        this.widthMm = source["widthMm"];
	        this.heightMm = source["heightMm"];
	    }
	}
	export class PaperSpecCatalog {
	    list: PaperSpec[];
	    default: string;
	    current: string;
	
	    static createFrom(source: any = {}) {
	        return new PaperSpecCatalog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], PaperSpec);
	        this.default = source["default"];
	        this.current = source["current"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PhotoSpec {
	    value: string;
	    title: string;
	    desc: string;
	    categories: string[];
	    keywords: string;
	    widthMm?: number;
	    heightMm?: number;
	    widthPx?: number;
	    heightPx?: number;
	    unit?: string;
	
	    static createFrom(source: any = {}) {
	        return new PhotoSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.value = source["value"];
	        this.title = source["title"];
	        this.desc = source["desc"];
	        this.categories = source["categories"];
	        this.keywords = source["keywords"];
	        this.widthMm = source["widthMm"];
	        this.heightMm = source["heightMm"];
	        this.widthPx = source["widthPx"];
	        this.heightPx = source["heightPx"];
	        this.unit = source["unit"];
	    }
	}
	export class SpecCategory {
	    key: string;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new SpecCategory(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	    }
	}
	export class PhotoSpecCatalog {
	    list: PhotoSpec[];
	    default: string;
	    current: string;
	    categories: SpecCategory[];
	
	    static createFrom(source: any = {}) {
	        return new PhotoSpecCatalog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.list = this.convertValues(source["list"], PhotoSpec);
	        this.default = source["default"];
	        this.current = source["current"];
	        this.categories = this.convertValues(source["categories"], SpecCategory);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class SpecQuery {
	    keyword: string;
	    category: string;
	
	    static createFrom(source: any = {}) {
	        return new SpecQuery(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.keyword = source["keyword"];
	        this.category = source["category"];
	    }
	}

}

