package devx

//import (
// "time"
//)

#Worker: {
	TaskQueue: string | *"default"
	Workflows: [name=string]:  #Workflow & {Name: name}
	Activities: [name=string]: #Activity & {Name: name}
}
