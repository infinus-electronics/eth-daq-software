export namespace server {
	
	export class BufferKey {
	    IP: string;
	    Port: number;
	
	    static createFrom(source: any = {}) {
	        return new BufferKey(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.IP = source["IP"];
	        this.Port = source["Port"];
	    }
	}

}

