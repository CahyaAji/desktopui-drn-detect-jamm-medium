export namespace main {
	
	export class JammerConfig {
	    address: string;
	    port: number;
	    threshold: number;
	
	    static createFrom(source: any = {}) {
	        return new JammerConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.port = source["port"];
	        this.threshold = source["threshold"];
	    }
	}
	export class MapsConfig {
	    latitude: number;
	    longitude: number;
	    zoom: number;
	    maptilerKey: string;
	
	    static createFrom(source: any = {}) {
	        return new MapsConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.latitude = source["latitude"];
	        this.longitude = source["longitude"];
	        this.zoom = source["zoom"];
	        this.maptilerKey = source["maptilerKey"];
	    }
	}
	export class DetectorConfig {
	    address: string;
	    port: number;
	
	    static createFrom(source: any = {}) {
	        return new DetectorConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.port = source["port"];
	    }
	}
	export class AppConfig {
	    version: number;
	    detector: DetectorConfig;
	    maps: MapsConfig;
	    jammer: JammerConfig;
	
	    static createFrom(source: any = {}) {
	        return new AppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.detector = this.convertValues(source["detector"], DetectorConfig);
	        this.maps = this.convertValues(source["maps"], MapsConfig);
	        this.jammer = this.convertValues(source["jammer"], JammerConfig);
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
	export class Detection {
	    id: number;
	    timestamp: number;
	    serialNumber: string;
	    model: string;
	    latitude: number;
	    longitude: number;
	    scanFreqMHz: number;
	    droneRange: number;
	    altitude: number;
	
	    static createFrom(source: any = {}) {
	        return new Detection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.serialNumber = source["serialNumber"];
	        this.model = source["model"];
	        this.latitude = source["latitude"];
	        this.longitude = source["longitude"];
	        this.scanFreqMHz = source["scanFreqMHz"];
	        this.droneRange = source["droneRange"];
	        this.altitude = source["altitude"];
	    }
	}
	
	

}

