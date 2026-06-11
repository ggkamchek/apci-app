export namespace fingerprint {
	
	export class Signals {
	    userAgent: string;
	    platform: string;
	    language: string;
	    screenWidth: number;
	    screenHeight: number;
	    colorDepth: number;
	    timeZone: string;
	    timezoneOffset: number;
	    canvasHash: string;
	    webGLVendor: string;
	    webGLRenderer: string;
	    audioContextHash: string;
	    fontsHash: string;
	
	    static createFrom(source: any = {}) {
	        return new Signals(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.userAgent = source["userAgent"];
	        this.platform = source["platform"];
	        this.language = source["language"];
	        this.screenWidth = source["screenWidth"];
	        this.screenHeight = source["screenHeight"];
	        this.colorDepth = source["colorDepth"];
	        this.timeZone = source["timeZone"];
	        this.timezoneOffset = source["timezoneOffset"];
	        this.canvasHash = source["canvasHash"];
	        this.webGLVendor = source["webGLVendor"];
	        this.webGLRenderer = source["webGLRenderer"];
	        this.audioContextHash = source["audioContextHash"];
	        this.fontsHash = source["fontsHash"];
	    }
	}

}

export namespace main {
	
	export class AuthResult {
	    success: boolean;
	    message: string;
	    userId?: string;
	    username?: string;
	    sessionId?: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.userId = source["userId"];
	        this.username = source["username"];
	        this.sessionId = source["sessionId"];
	    }
	}

}

