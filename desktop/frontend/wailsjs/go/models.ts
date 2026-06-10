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

