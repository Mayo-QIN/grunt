import requests
import os


class job(object):
    def __init__(self,endpoint):
        self.endpoint = endpoint
        self.files = {}
        self.data = {}

    def call(self, data, files):
        if len(files) == 0:
            files['bogus_empty_name'] = ("", 'content');
        resp = requests.post(self.endpoint.url(), files=files, data=data)
        print resp.text
        self.json = resp.json()
        print ( self.json )
        self.uuid = self.json['uuid']

    def url(self):
        return self.endpoint.address + "/rest/job/" + self.json['uuid']
        
    def status(self):
        self.job_status = requests.get(self.url()).json()
        return self.job_status

    def wait(self):
		self.job_status = requests.get(self.endpoint.address + "/rest/job/wait/" + self.json['uuid'])
		return self.job_status

    def save_all(self,directory):
        for k in self.endpoint.outputs.keys():
            self.save_output(k,directory)
    
    def save_output(self,key,directory):
        print "Saving " + key + " to " + directory
        if key in self.data.keys():
            v = self.data[key]
            try:
                r1 = requests.get( self.url() + "/file/" + key )
                print r1
                with open( os.path.join(directory,v), "wb") as code:
                    code.write(r1.content)
                print 'done'
            except Exception, e: print e

       
class endpoint(object):
    """
    Manage an endpoint
    """     
    def __init__(self, address, endpoint):
        self.address = address
        self.endpoint = endpoint
        self.json = requests.get(self.url()).json()
    def parameters(self):
        return self.json['parameters']
    def inputs(self):
        return self.json['input_files']
    def outputs(self):
        return self.json['output_files']
    def url(self):
        return self.address+"/rest/service/"+self.endpoint

    def __call__(self,**kwargs):
        j = job(self)
        print (kwargs)
        for k,v in kwargs.iteritems():
            if k in self.parameters():
                j.data[k] = v
            if k in self.inputs():
                j.files[k] = open(v,'rb')
            if k in self.outputs():
                j.data[k] = v
        j.call(j.data, j.files)
        return j


class grunt(object):
    """
    A class that manages the comunication with the web services offered by Grunt.
    """
    def __init__(self, address,user='none', password='none'):
        self.address = address
        # Cache the services
        self.services = {}
        self.services_json = requests.get(self.address + "/rest/service").json()
        for service in self.services_json["services"]:
            self.services[service["end_point"]] = endpoint(self.address, service["end_point"])

    def __getattr__(self,key):
        return self.services[key]
            
	def description(self):
		print "I'm a class for this server %s and specifically this %s end point." % (self.address, self.service)
 
	def submitjob(self):
		r = requests.post(self.servicecontactlocation, files=self.files, data=self.param)
		self.r=r
		return 0

	def waitforcompletion(self):
		robj=self.r
		ConnObject=robj.json()
		status = requests.get(self.address+'/rest/job/wait/'+ConnObject.get('uuid'))
		self.status =status
		return 0      

	def download(self):
		robj=self.r
		ConnObject=robj.json()
		filespassed=self.param	 	
	 	for k, v in filespassed.iteritems():
	 		print k,v
	 		try:
				r1 = requests.get( self.address+'/rest/job/'+ConnObject.get('uuid')+'/file/'+k)
				with open( self.storelocation+v, "wb") as code:
					code.write(r1.content)
				print 'done'
			except Exception, e: print e
			
			return 0  


# n4 = webappserver(address,param,files,storelocation, service)

if __name__ == "__main__":

    g = grunt("http://localhost:9901")
    j = g.copy(input="README.md",output="Copy of README.md")
    j.wait()

    g.echo(Message='Hi from grunt')
    j.wait()
    j.save_output("output", "/tmp/")

    
    # e = endpoint("http://localhost:9901", "copy")
    # j = e(input="README.md",output="Copy of README.md")
    # j.wait()
