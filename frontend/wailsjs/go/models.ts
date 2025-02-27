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
	export class IPConnection {
	    ActivePorts: Record<number, boolean>;
	    TotalBytes: number;
	    UUID: string;
	
	    static createFrom(source: any = {}) {
	        return new IPConnection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ActivePorts = source["ActivePorts"];
	        this.TotalBytes = source["TotalBytes"];
	        this.UUID = source["UUID"];
	    }
	}

}

