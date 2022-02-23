package view

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)
var(

	upgrader = websocket.Upgrader{}
	port string = ":5555"
	rootws string = "/ws"
	contenido string  = "<html></html>"
	rootserv string = "http://localhost" + port + "/"
	com *websocket.Conn
	conection bool = false
	done bool = false
	methods []method
	windowRef []Element
	windowLoad = false
	event *Event
)
// types
type method struct{
   name string
   function func()
}
type Element struct {
   Ref string `json:"ref"`
   Id string `json:"id"`
   ClassName string `json:"className"`
   InnerHTML string `json:"innerHTML"`
   TagName string `json:"tagName"`
   ParentNode string `json:"parentNode"`
   Children string	`json:"children"`
   Value string	`json:"value"`
   Position int
}
type smsJsons struct{
   Type string `json:"type"`
   Ref string `json:"ref"`
   Body string `json:"body"`
   Value string `json:"value"`
   Field string `json:"field"`
   Event string `json:"event"`
   Name string `json:"name"`
}
type Component struct{
	Name string 
	Render string 
	Style string
	Action func()
	Childrens []*Component
}
type Event struct{
	Type string `json:"type"`
	Target string `json:"target"`
}
// utils 
func Error(err error){
	if err != nil{
		fmt.Println(err)
	}
}
func Log( s ...interface{}){
	fmt.Println( s... )
}
func ToFirstUpperCase(str string)string{
	str = strings.ToTitle(str[:1]) + str[1:]
	return str
}
// eventos
func onWindowLoad(call func()){
	for{
		if windowLoad{
			call()
			return 
		}
	}
}
func OnWait(){
	for {
		if done == true {
			Log("Cerando...")
			return 
		}
	}
}
// control window
func isWindows(browser string) bool {

	if browser == "chrome" {

		cmd := exec.Command("c:/Program Files (x86)/Google/Chrome/Application/./chrome", "--app="+fmt.Sprintf("%s", rootserv))
		if err := cmd.Start(); err != nil { // Ejecutar comando

			log.Fatal(err)
			return false

		}
	} else if browser == "firefox" {
	} else if browser == "edge" {
	}
	return true
}
func js() string{
	js:= `
	<script>

	const ws = new WebSocket ("ws://localhost`+ port + rootws +`")

	ws.onopen = (e)=>{

		console.log("conectado")
		ws.send("ok")

	}

	ws.onmessage = (e)=>{
		console.log(e.data)
		let data = JSON.parse( e.data )

		if ( data.type == "bind" ){
			isBind( data )
			return
		}
		if ( data.type == "eval" ){
			isEval( data )
			return 
		}
		if ( data.type == "window" ){
			isWindow( data )
			return 
		}
	}

	function isBind( data ){

		window[data.name] = ()=>{
			if ( event.type != "dragstart"){
				event.preventDefault()
			}
			ws.send( JSON.stringify({type:"event", name:data.name ,event:JSON.stringify({type:event.type,target:event.target.id})}) )
		}
	}

	function isEval( data ){

		let res = eval( data.js )

		if ( typeof res != "string" ){
			res = JSON.stringify( res )
		}
		if ( data.id ){
			res = JSON.stringify( {id : data.id , body: res} )
		}
		if ( res != undefined ){
			ws.send( res )
		}
	}

	function isWindow( data ){

		let res = eval( data.js )

		res.forEach(item=>{
			if ( item.tagName != "SCRIPT" && item.tagName != "BODY" && item.tagName != "STYLE"){
				uploadValue( item )
			}
			ws.send( JSON.stringify( {type:"window", ref:item.ref , field :(res.length-1).toString() ,body:JSON.stringify(item)} )
		)
		})
	}

	window.addEventListener('beforeunload', (e)=>{
		e.preventDefault()
		ws.send("close")
	})

	window.uploadValue = ( data )=>{
		let res = data
		let ele = document.querySelector('.' + data.ref )

		ele.addEventListener('change',()=>{
				res.value = ele.value.toString()
				ws.send(JSON.stringify({type:"upload", Ref:data.ref , value : res.value , body :JSON.stringify(res)}))
		})
	}

	window.getElements = ()=>{

			let win = document.body.querySelectorAll("*")
			const attrs = ["ref","id","className","innerHTML","tagName","parentNode","value"]
			
			win = Array.from( win ).map(( item , index )=>{
				item.classList.add( "socket_"+ index )
			    let obj ={}
			    for (let i in item ){
			        attrs.forEach(att =>{
			            if ( att == i ){
			                if (i == "parentNode"){
			                    obj[i] = JSON.stringify({tagName : item[i].tagName , id : item[i].id , class: item[i].className})
			                }else{
			                    obj[i] = item[i]
			                }
			            }
			        })
			    }
			    obj["ref"] = "socket_"+ index
			    return obj
			})
			return win
	}
	</script>
	`
	return js
}
func New(browser string, title string, content Component){

	if title != ""{
		content.Render = strings.Replace(content.Render,"<html>","<html><title>"+ title +"</title>", 1)
	}
	if content.Style != ""{
		content.Render = Build(content)
		content.Render = strings.Replace( content.Render,"<body>","<body><style>"+ content.Style +"</style>", 1)
		if len(content.Childrens) > 0 {
			for _, child := range content.Childrens{
				content.Render = strings.Replace(content.Render,"</style>",child.Style + "</style>",1)
			}
		}
	}
	// inyect js
	content.Render = strings.Replace(content.Render,"<body>" , "<body>"+ js() , 1)
	contenido = content.Render

	if runtime.GOOS == "windows" {
		isWindows("chrome")
	}
	// start server and window
	go newServer()
	// obtengo elementos del objeto window
	getWindow()
	go onWindowLoad(func(){
		// ejecutar action de todos los componentes
		content.Action()
		for _, child := range content.Childrens{
			child.Action()
		}
	})
}
// building html con componentes 
func Build( ele Component )string{
	for _, child := range ele.Childrens{
		if strings.Contains(ele.Render,"</"+child.Name+">"){
			ele.Render = strings.Replace(ele.Render,"</"+child.Name+">",child.Render,1)
		}
	}
	//fmt.Println(ele.Render)
	return ele.Render
}
// comunication
func send( sms string ){
	//tipo := "undefined"
	//if strings.Contains(sms,"bind"){ tipo = "is funtion" }else
	//if strings.Contains(sms,"addEventListener"){ tipo = "is event" }else
	//if strings.Contains(sms,"eval"){ tipo = "is evaluation" }
	for {
		if conection{
			_ = com.WriteMessage(1,[]byte(sms))
			//Log("message sent : " , sms  ,"   ", tipo)
			return
		}
	}
}
func upload(s string ){

	var obj smsJsons
	var ele Element

	err := json.Unmarshal([]byte( s ),&obj)
		Error( err )
	err = json.Unmarshal([]byte( obj.Body ),&ele)
		Error( err )
	ele.Ref = obj.Ref

	
		for  i , e := range windowRef{
			if e.Ref == ele.Ref {
				windowRef[i] = ele
				windowRef[i].Position = i
			}
		}	
}
// obtener referencia de windowJs
func getWindow(){
	js := `{"type":"window","js": "getElements()"}`
	eval( js )
}
// functions server 
func reciver(w http.ResponseWriter, r *http.Request) {
	com, _ = upgrader.Upgrade(w, r, nil)
	defer com.Close()

	for {
		_, tempSMS, _ := com.ReadMessage()
		// Receive message
		evalOptions(string(tempSMS))
	}
}
func serv(w http.ResponseWriter, r *http.Request){
	fmt.Fprint(w , contenido )
}
func newServer() {
	http.HandleFunc(rootws, reciver)
	http.HandleFunc("/" , serv )
	log.Fatal(http.ListenAndServe(port, nil))
}
// add methods 
func AddMethod( name string , f func()){
	methods = append(methods, method{ name , f })
}
func evalMethods( sms string )bool{
	var Json smsJsons
	json.Unmarshal([]byte(sms), &Json)
	json.Unmarshal([]byte(Json.Event),&event)
	for _,v := range methods{
		if v.name == Json.Name {
			v.function()
			return true
		}
	}
	return false
}
// evaluation options reciver sw
func evalOptions(sms string){

	if sms == "ok"{
		ok()
		return
	}
	if sms == "close"{
		close( sms )
		return
	}
	if strings.Contains( sms , "upload"){
		upload( sms )
		return
	}
	if strings.Contains( sms , "window"){
		window( sms )
		return
	}
	evalMethods(sms)
}
func ok(){
	conection = true
	Log("conection is OK!")
}
func close( sms string ){
	done = true
}
func window( sms string ){

	var smsJson smsJsons 
	var window Element

	err := json.Unmarshal([]byte(sms),&smsJson)
		Error( err )
	err = json.Unmarshal([]byte(smsJson.Body),&window)
		Error( err )

	windowRef = append( windowRef , window )
	length ,_ := strconv.Atoi(smsJson.Field)

	if len(windowRef) == length {
		windowLoad = true
		Log("window is ok!")
	}
}
// bind js
func Bind(name string , f func()){
	eval(`{"type":"bind","name":"`+ name +`"}`)
	AddMethod( name , f )
	return 
}
func eval( s string ){
	for {
		if conection{
			_ = com.WriteMessage(1,[]byte(s))
			//Log("message sent : " , s )
			return
		}
	}
}
// animation 
func getBoince()string{
	return `@keyframes bounceIn{0%{opacity: 0;transform: scale(0.3) translate3d(0,0,0);}50%{opacity: 0.9;transform: scale(1.1);}80%{opacity: 1;transform: scale(0.89);}100%{opacity: 1;transform: scale(1) translate3d(0,0,0);}}`
}
// methods ui
func getNameComponent(str string )string {
	res := strings.Index(str,"class")
	if res != -1 {
		str = str[res:res+20]
		str = strings.Split(str,"=")[1]
		str = strings.Replace(str, `"`,"",1)
		res:= strings.Index(str,`"`)
		str = str[:res]
		str = strings.TrimSpace(str)
		str = ToFirstUpperCase(str)
	}else{
		fmt.Println("error al obtener el nombre ")
		str = ""
	}
	return str
}
func AddChilds(childs ...*Component)[]*Component{
	for i,v :=range childs{
		
		childs[i].Name = getNameComponent(v.Render)

	}
	return childs
}
// methos js 
func Selector( query string )(res *Element){

	for i , v := range windowRef{
		if strings.ToLower(v.TagName) == query || strings.Contains( v.ClassName , strings.Replace(query,".","",1) )  || "#"+v.Id == query{
			res =  &windowRef[i]
			return 
		}
	}
	null := Element{ClassName: "<nil>"}
	res = &null
	return 
}
func SelectorById(query string)(res *Element){
	return Selector("#"+query)
}
func InnerToggle( a *Element , b *Element ){
	c:= a.InnerHTML 
	a.SetInnerHTML(b.InnerHTML) 
	b.SetInnerHTML(c)
}
// parent
type parent struct{
	TagName string `json:"tagName"`
	Id string `json:"id"`
	Class string `json:"class"`
}
func (this *Element) GetParent()*parent{
	var vParent parent
	json.Unmarshal([]byte(this.ParentNode),&vParent)
	return &vParent
}
func ( this *Element ) AppendChild( child *Element ){
	this.InnerHTML += SelectorById(child.GetParent().Id).InnerHTML
	fmt.Println(this.Id ,":",this.InnerHTML)
	js := `{"type":"eval","js":"document.getElementById('`+ this.Id +`').appendChild(document.getElementById('`+ child.Id +`'))"}`
	eval( js )
}
func ( this *Element ) GetFirstChild()*Element{
	i := strings.LastIndex(this.InnerHTML,"id")
	if i != -1{
		str := strings.Split(this.InnerHTML[i:i+20],"=")[1]
		str = strings.TrimSpace(strings.Split(strings.Replace(str,`"`,"",1),`"`)[0])
		fmt.Println(str)
		return SelectorById(str)
	}else{
		return &Element{}
	}
}
func ( this *Element ) ChangeChild( ele *Element ){
	childA := this.GetFirstChild()
	childB := ele.GetFirstChild()
	cloneA := childA
	cloneB := childB
	childB = cloneA
	childA = cloneB
	childInnerA := this.InnerHTML
	childInnerB := ele.InnerHTML
	this.InnerHTML = childInnerB
	ele.InnerHTML = childInnerA
	js := `{"type":"eval","js":"let childA = document.getElementById('`+this.Id+`').innerHTML;let childB = document.getElementById('`+ele.Id+`').innerHTML;document.getElementById('`+ele.Id+`').innerHTML = childA;document.getElementById('`+this.Id+`').innerHTML = childB;"}`
	eval( js )
}
func ( this *Element ) SetInnerHTML( s interface{} )*Element{
	this.InnerHTML = strings.TrimSpace(fmt.Sprint(s))

	js := `{"type":"eval","js":"document.querySelector('.`+ this.Ref +`').innerHTML ='`+ this.InnerHTML +`'"}`
	eval ( js )
	return this
}
func ( this *Element ) SetValue( s string )*Element{
	this.Value = s
	js := `{"type":"eval","js":"document.querySelector('.`+ this.Ref +`').value ='`+ this.Value +`'"}`
	eval ( js )
	return this
}
func ( this *Element ) SetClassName( s string )*Element{
	this.ClassName = s
	js := `{"type":"eval","js":"document.querySelector('.`+ this.Ref +`').className ='`+ this.ClassName +`'"}`
	eval ( js )
	js = `{"type":"eval","js":"document.querySelector('.`+ this.ClassName +`').classList.add('`+ this.Ref +`')"}`
	eval( js )
	return this
}
func ( this *Element ) SetId( s string )*Element{
	this.Id = s
	js := `{"type":"eval","js":"document.querySelector('.`+ this.Ref +`').id ='`+ this.Id +`'"}`
	eval ( js )
	return this
}
func ( this *Element ) SetAttribute( s string , v string )*Element{
	js := `{"type":"eval","js":"document.querySelector('.`+ this.Ref +`').setAttribute('`+ s +`','`+ v +`')"}`
	eval ( js )
	return this
}
func ( this *Element ) AddEventListener( tipo string , name string , f func()){
	js := `{"type":"eval","js":"document.querySelector('.`+ this.Ref +`').addEventListener('`+ tipo +`',` + name + `)"}`
	Bind( name , f )
	eval( js )
}
func ( this *Element ) SetStyles( n string , v string ){
	js := `{"type":"eval","js":"document.querySelector('.`+ this.Ref +`').style.`+ n +`= '`+ v +`'"}`
	eval( js )
}
func ( this *Element ) Remove(){
	js := `{"type":"eval","js":"document.querySelector('.`+ this.Ref +`').remove()"}`
	eval( js )
}
func ( this *Element ) Close( op string  , t float32 ){

	var ani string
	var keyframe bool = false
	var js string
	var f string

	if op == "dispel"{
		ani = `ele.style.opacity='0';`
	}
	if op == "rotate_x"{
		ani = `ele.style.transform ='rotateX(90deg)';ele.style.opacity='0';`
	}
	if op == "rotate_y"{
		ani = `ele.style.transform ='rotateY(90deg)';ele.style.opacity='0';`
	}
	if op == "rotate_z"{
		ani = `ele.style.transform ='rotateZ(180deg)';ele.style.opacity = '0'`
	}
	if op == "boince"{
		keyframe = true
	}
	if !keyframe{
		f = `(()=>{let ele = document.querySelector('.`+ this.Ref +`');ele.style.transition = 'all `+ fmt.Sprint(t) +`s';`+ ani +`})()`
		js = `{"type":"eval","js":"`+ f +`"}`
	}else{
		f = `document.querySelector('style').innerHTML += '`+ getBoince() +`';let ele = document.querySelector('.`+ this.Ref +`');ele.style.opacity = '0';ele.style.animation = 'bounceIn `+fmt.Sprint(t)  +`s reverse'`
		js = `{"type":"eval","js":"`+ f +`"}`
	}
	eval( js )
}
func ( this *Element ) Open( op string  , t float32 ){

	var ani string
	var keyframe bool = false
	var js string
	var f string

	if op == "dispel"{
		ani = `ele.style.opacity='1';`
	}
	if op == "rotate_x"{
		ani = `ele.style.transform ='rotateX(0)';ele.style.opacity='1';`
	}
	if op == "rotate_y"{
		ani = `ele.style.transform ='rotateY(0)';ele.style.opacity='1';`
	}
	if op == "rotate_z"{
		ani = `ele.style.transform ='rotateZ(0)';ele.style.opacity='1'`
	}	
	if op == "boince"{
		keyframe = true
	}
	if !keyframe{
		f = `(()=>{let ele = document.querySelector('.`+ this.Ref +`');ele.style.transition = 'all `+ fmt.Sprint(t) +`s';`+ ani +`})()`
		js = `{"type":"eval","js":"`+ f +`"}`
	}else{
		f = `document.querySelector('style').innerHTML += '`+ getBoince() +`';let ele = document.querySelector('.`+ this.Ref +`');ele.style.opacity = '1';ele.style.animation = 'bounceIn `+ fmt.Sprint(t)  +`s'`
		js = `{"type":"eval","js":"`+ f +`"}`
	}
	eval( js )
}
func ( this *Element ) ToggleAni( op string , t float32 ){
	this.Close( op , t )
	time.Sleep(1*time.Second)
	this.Open( op , t )
}
// chrono
type alarm struct{
	minute int 
	second int 
}
type chrono struct{
	ele *Element
	ala alarm
	call func()
}
func ( this *Element ) IsChrono()( *chrono ){
	var chr chrono
	chr.ele = this
	return &chr
}
func ( chr *chrono ) Run(  minute int , second int  ){
	second ++
	if minute > 60 { minute = 60 }
	if second > 60 { second = 60 }
	go func(){
		for{
			if second == chr.ala.second+1 && minute == chr.ala.minute{ chr.call() }
			if second > 60 { second = 0 ; minute++ }
	
			time.Sleep(1*time.Second)
	
			if minute < 10 && second < 10 { 
				chr.ele.SetInnerHTML(fmt.Sprint("0",minute,":","0",second))
			}else if minute < 10 && second >= 10 { 
				chr.ele.SetInnerHTML(fmt.Sprint("0",minute,":",second))
			}else{ 
				chr.ele.SetInnerHTML(fmt.Sprint(minute,":",second)) }
			second++
		}

	}()
}
func ( chr *chrono ) SetAlarm( minute int , second int , call func())*chrono{
	chr.ala.minute = minute
	chr.ala.second = second
	chr.call = call
	return chr
}
// contador 
type counter struct{
	ele *Element
	Max int
	Min int
}
func (this *Element) IsCounter()*counter{
	cont := counter{
		ele : this,
		Max:1000000,
		Min:-1000000,
	}
	return &cont 
}
func ( this *counter ) SetMax( max int)*counter {
	this.Max = max
	return this
}
func ( this *counter ) SetMin( min int )*counter{
	this.Min = min
	return this
}
func ( this *counter ) Increment(){
	inner ,_:= strconv.Atoi(this.ele.InnerHTML)
	if inner < this.Max{
		this.ele.SetInnerHTML(inner + 1 )
	}
}
func ( this *counter ) Decrement(){
	inner,_:= strconv.Atoi(this.ele.InnerHTML)
	if inner > this.Min{ 
		this.ele.SetInnerHTML(inner - 1)
	}
}
// event 
func GetEvent()*Event{
	return event
}














