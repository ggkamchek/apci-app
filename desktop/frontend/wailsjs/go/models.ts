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
	export class E2EInboxMessage {
	    messageId: string;
	    chatId: string;
	    senderId: string;
	    plaintext: string;
	    sentAtUnix: number;
	
	    static createFrom(source: any = {}) {
	        return new E2EInboxMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.messageId = source["messageId"];
	        this.chatId = source["chatId"];
	        this.senderId = source["senderId"];
	        this.plaintext = source["plaintext"];
	        this.sentAtUnix = source["sentAtUnix"];
	    }
	}
	export class E2EInboxResult {
	    success: boolean;
	    message: string;
	    messages?: E2EInboxMessage[];
	
	    static createFrom(source: any = {}) {
	        return new E2EInboxResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.messages = this.convertValues(source["messages"], E2EInboxMessage);
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
	export class E2EResult {
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new E2EResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}
	export class E2ESendResult {
	    success: boolean;
	    message: string;
	    messageId?: string;
	
	    static createFrom(source: any = {}) {
	        return new E2ESendResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.message = source["message"];
	        this.messageId = source["messageId"];
	    }
	}

}

