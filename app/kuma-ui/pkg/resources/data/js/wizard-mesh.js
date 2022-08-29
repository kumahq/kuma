"use strict";(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[708],{88523:function(e,a,t){var n=t(73570),l=t.n(n);a["Z"]={methods:{formatForCLI(e,a='" | kumactl apply -f -'){const t='echo "',n=l()(e);return`${t}${n}${a}`}}}},27545:function(e){e.exports={mtls:{enabledBackend:null,backends:[]},tracing:{defaultBackend:null,backends:[{name:null,type:null}]},logging:{backends:[{name:null,format:'{ "destination": "%KUMA_DESTINATION_SERVICE%", "destinationAddress": "%UPSTREAM_LOCAL_ADDRESS%", "source": "%KUMA_SOURCE_SERVICE%", "sourceAddress": "%KUMA_SOURCE_ADDRESS%", "bytesReceived": "%BYTES_RECEIVED%", "bytesSent": "%BYTES_SENT%"}',type:null}]},metrics:{enabledBackend:null,backends:[{name:null,type:null}]}}},87287:function(e,a,t){t.r(a),t.d(a,{default:function(){return Se}});var n=t(70821);const l={class:"wizard"},i={class:"wizard__content"},s=(0,n._)("code",null,"kubectl",-1),d=(0,n.Uk)(" (if you are running in Kubernetes mode) or "),o=(0,n._)("code",null,"kumactl",-1),r=(0,n.Uk)(" / API (if you are running in Universal mode). "),c=(0,n._)("h3",null," To get started, please fill in the following information: ",-1),m={class:"k-input-label mx-2"},u=(0,n._)("span",null,"Disabled",-1),g={class:"k-input-label mx-2"},p=(0,n._)("span",null,"Enabled",-1),h=(0,n._)("option",{value:"builtin"}," builtin ",-1),b=(0,n._)("option",{value:"provided"}," provided ",-1),y=[h,b],k=(0,n._)("p",{class:"help"}," If you've enabled mTLS, you must select a CA. ",-1),w=(0,n._)("h3",null," Setup Logging ",-1),f=(0,n._)("p",null,' You can setup as many logging backends as you need that you can later use to log traffic via the "TrafficLog" policy. In this wizard, we allow you to configure one backend, but you can add more manually if you wish. ',-1),_={class:"k-input-label mx-2"},v=(0,n._)("span",null,"Disabled",-1),E={class:"k-input-label mx-2"},T=(0,n._)("span",null,"Enabled",-1),S={key:1},M=(0,n._)("option",{value:"tcp"}," TCP ",-1),U=(0,n._)("option",{value:"file"}," File ",-1),C=[M,U],L=(0,n._)("h3",null," Setup Tracing ",-1),N=(0,n._)("p",null,' You can setup as many tracing backends as you need that you can later use to log traffic via the "TrafficTrace" policy. In this wizard we allow you to configure one backend, but you can add more manually as you wish. ',-1),A={class:"k-input-label mx-2"},D=(0,n._)("span",null,"Disabled",-1),V={class:"k-input-label mx-2"},B=(0,n._)("span",null,"Enabled",-1),x=(0,n._)("option",{value:"zipkin"}," Zipkin ",-1),I=[x],R=(0,n._)("h3",null," Setup Metrics ",-1),W=(0,n._)("p",null," You can expose metrics from every data-plane on a configurable path and port that a metrics service, like Prometheus, can use to fetch them. ",-1),O={class:"k-input-label mx-2"},q=(0,n._)("span",null,"Disabled",-1),P={class:"k-input-label mx-2"},j=(0,n._)("span",null,"Enabled",-1),z=(0,n._)("option",{value:"prometheus"}," Prometheus ",-1),G=[z],F={key:0},K={key:0},Z=(0,n._)("h3",null," Install a new Mesh ",-1),Y=(0,n._)("h3",null,"Searching…",-1),$=(0,n._)("p",null,"We are looking for your mesh.",-1),H=(0,n._)("h3",null,"Done!",-1),J=(0,n.Uk)(" Your Mesh "),Q={key:0},X=(0,n.Uk)(" was found! "),ee=(0,n.Uk)(" See Meshes "),ae=(0,n._)("h3",null,"Mesh not found",-1),te=(0,n._)("p",null,"We were unable to find your mesh.",-1),ne=(0,n._)("p",null," You haven't filled any data out yet! Please return to the first step and fill out your information. ",-1),le=(0,n._)("h3",null,"Mesh",-1),ie=["href"],se=(0,n._)("h3",null,"Did You Know?",-1),de=(0,n._)("p",null," As you know, the GUI is read-only, but it will be providing instructions to create a new Mesh and verify everything worked well. ",-1);function oe(e,a,t,h,b,M){const U=(0,n.up)("KAlert"),x=(0,n.up)("FormFragment"),z=(0,n.up)("KCard"),oe=(0,n.up)("CodeView"),re=(0,n.up)("TabsWidget"),ce=(0,n.up)("KButton"),me=(0,n.up)("EntityScanner"),ue=(0,n.up)("StepSkeleton");return(0,n.wg)(),(0,n.iD)("div",l,[(0,n._)("div",i,[(0,n.Wm)(ue,{steps:b.steps,"sidebar-content":b.sidebarContent,"footer-enabled":!1===b.hideScannerSiblings,"next-disabled":M.nextDisabled},{general:(0,n.w5)((()=>[(0,n._)("p",null," Welcome to the wizard for creating a new Mesh resource in "+(0,n.zw)(b.productName)+". We will be providing you with a few steps that will get you started. ",1),(0,n._)("p",null,[(0,n.Uk)(" As you know, the "+(0,n.zw)(b.productName)+" GUI is read-only, so at the end of this wizard we will be generating the configuration that you can apply with either ",1),s,d,o,r]),c,(0,n.Wm)(z,{class:"my-6 k-card--small",title:"Mesh Information","has-shadow":""},{body:(0,n.w5)((()=>[(0,n.Wm)(x,{title:"Mesh name","for-attr":"mesh-name"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"mesh-name","onUpdate:modelValue":a[0]||(a[0]=e=>b.validate.meshName=e),type:"text",class:"k-input w-100",placeholder:"your-mesh-name",required:""},null,512),[[n.nr,b.validate.meshName]]),b.vmsg.meshName?((0,n.wg)(),(0,n.j4)(U,{key:0,appearance:"danger",size:"small","alert-message":b.vmsg.meshName},null,8,["alert-message"])):(0,n.kq)("",!0)])),_:1}),(0,n.Wm)(x,{title:"Mutual TLS"},{default:(0,n.w5)((()=>[(0,n._)("label",m,[(0,n.wy)((0,n._)("input",{ref:"mtlsDisabled","onUpdate:modelValue":a[1]||(a[1]=e=>b.validate.mtlsEnabled=e),value:"disabled",name:"mtls",type:"radio",class:"k-input mr-2"},null,512),[[n.G2,b.validate.mtlsEnabled]]),u]),(0,n._)("label",g,[(0,n.wy)((0,n._)("input",{id:"mtls-enabled","onUpdate:modelValue":a[2]||(a[2]=e=>b.validate.mtlsEnabled=e),value:"enabled",name:"mtls",type:"radio",class:"k-input mr-2"},null,512),[[n.G2,b.validate.mtlsEnabled]]),p])])),_:1}),"enabled"===b.validate.mtlsEnabled?((0,n.wg)(),(0,n.j4)(x,{key:0,title:"Certificate name","for-attr":"certificate-name"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"certificate-name","onUpdate:modelValue":a[3]||(a[3]=e=>b.validate.meshCAName=e),type:"text",class:"k-input w-100",placeholder:"your-certificate-name"},null,512),[[n.nr,b.validate.meshCAName]])])),_:1})):(0,n.kq)("",!0),"enabled"===b.validate.mtlsEnabled?((0,n.wg)(),(0,n.j4)(x,{key:1,title:"Certificate Authority","for-attr":"certificate-authority"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("select",{id:"certificate-authority","onUpdate:modelValue":a[4]||(a[4]=e=>b.validate.meshCA=e),class:"k-input w-100",name:"certificate-authority"},y,512),[[n.bM,b.validate.meshCA]]),k])),_:1})):(0,n.kq)("",!0)])),_:1})])),logging:(0,n.w5)((()=>[w,f,(0,n.Wm)(z,{class:"my-6 k-card--small",title:"Logging Configuration","has-shadow":""},{body:(0,n.w5)((()=>[(0,n.Wm)(x,{title:"Logging"},{default:(0,n.w5)((()=>[(0,n._)("label",_,[(0,n.wy)((0,n._)("input",{id:"logging-disabled","onUpdate:modelValue":a[5]||(a[5]=e=>b.validate.loggingEnabled=e),value:"disabled",name:"logging",type:"radio",class:"k-input mr-2"},null,512),[[n.G2,b.validate.loggingEnabled]]),v]),(0,n._)("label",E,[(0,n.wy)((0,n._)("input",{id:"logging-enabled","onUpdate:modelValue":a[6]||(a[6]=e=>b.validate.loggingEnabled=e),value:"enabled",name:"logging",type:"radio",class:"k-input mr-2"},null,512),[[n.G2,b.validate.loggingEnabled]]),T])])),_:1}),"enabled"===b.validate.loggingEnabled?((0,n.wg)(),(0,n.j4)(x,{key:0,title:"Backend name","for-attr":"backend-name"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"backend-name","onUpdate:modelValue":a[7]||(a[7]=e=>b.validate.meshLoggingBackend=e),type:"text",class:"k-input w-100",placeholder:"your-backend-name"},null,512),[[n.nr,b.validate.meshLoggingBackend]])])),_:1})):(0,n.kq)("",!0),"enabled"===b.validate.loggingEnabled?((0,n.wg)(),(0,n.iD)("div",S,[(0,n.Wm)(x,{title:"Type"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("select",{id:"logging-type",ref:"loggingTypeSelect","onUpdate:modelValue":a[8]||(a[8]=e=>b.validate.loggingType=e),class:"k-input w-100",name:"logging-type"},C,512),[[n.bM,b.validate.loggingType]])])),_:1}),"file"===b.validate.loggingType?((0,n.wg)(),(0,n.j4)(x,{key:0,title:"Path","for-attr":"backend-address"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"backend-address","onUpdate:modelValue":a[9]||(a[9]=e=>b.validate.meshLoggingPath=e),type:"text",class:"k-input w-100"},null,512),[[n.nr,b.validate.meshLoggingPath]])])),_:1})):(0,n.kq)("",!0),"tcp"===b.validate.loggingType?((0,n.wg)(),(0,n.j4)(x,{key:1,title:"Address","for-attr":"backend-address"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"backend-address","onUpdate:modelValue":a[10]||(a[10]=e=>b.validate.meshLoggingAddress=e),type:"text",class:"k-input w-100"},null,512),[[n.nr,b.validate.meshLoggingAddress]])])),_:1})):(0,n.kq)("",!0),(0,n.Wm)(x,{title:"Format","for-attr":"backend-format"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("textarea",{id:"backend-format","onUpdate:modelValue":a[11]||(a[11]=e=>b.validate.meshLoggingBackendFormat=e),class:"k-input w-100 code-sample",rows:"12"},null,512),[[n.nr,b.validate.meshLoggingBackendFormat]])])),_:1})])):(0,n.kq)("",!0)])),_:1})])),tracing:(0,n.w5)((()=>[L,N,(0,n.Wm)(z,{class:"my-6 k-card--small",title:"Tracing Configuration","has-shadow":""},{body:(0,n.w5)((()=>[(0,n.Wm)(x,{title:"Tracing"},{default:(0,n.w5)((()=>[(0,n._)("label",A,[(0,n.wy)((0,n._)("input",{id:"tracing-disabled","onUpdate:modelValue":a[12]||(a[12]=e=>b.validate.tracingEnabled=e),value:"disabled",name:"tracing",type:"radio",class:"k-input mr-2"},null,512),[[n.G2,b.validate.tracingEnabled]]),D]),(0,n._)("label",V,[(0,n.wy)((0,n._)("input",{id:"tracing-enabled","onUpdate:modelValue":a[13]||(a[13]=e=>b.validate.tracingEnabled=e),value:"enabled",name:"tracing",type:"radio",class:"k-input mr-2"},null,512),[[n.G2,b.validate.tracingEnabled]]),B])])),_:1}),"enabled"===b.validate.tracingEnabled?((0,n.wg)(),(0,n.j4)(x,{key:0,title:"Backend name","for-attr":"tracing-backend-name"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"tracing-backend-name","onUpdate:modelValue":a[14]||(a[14]=e=>b.validate.meshTracingBackend=e),type:"text",class:"k-input w-100",placeholder:"your-tracing-backend-name"},null,512),[[n.nr,b.validate.meshTracingBackend]])])),_:1})):(0,n.kq)("",!0),"enabled"===b.validate.tracingEnabled?((0,n.wg)(),(0,n.j4)(x,{key:1,title:"Type","for-attr":"tracing-type"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("select",{id:"tracing-type","onUpdate:modelValue":a[15]||(a[15]=e=>b.validate.meshTracingType=e),class:"k-input w-100",name:"tracing-type"},I,512),[[n.bM,b.validate.meshTracingType]])])),_:1})):(0,n.kq)("",!0),"enabled"===b.validate.tracingEnabled?((0,n.wg)(),(0,n.j4)(x,{key:2,title:"Sampling","for-attr":"tracing-sampling"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"tracing-sampling","onUpdate:modelValue":a[16]||(a[16]=e=>b.validate.meshTracingSampling=e),type:"number",class:"k-input w-100",step:"0.1",min:"0",max:"100"},null,512),[[n.nr,b.validate.meshTracingSampling]])])),_:1})):(0,n.kq)("",!0),"enabled"===b.validate.tracingEnabled?((0,n.wg)(),(0,n.j4)(x,{key:3,title:"URL","for-attr":"tracing-zipkin-url"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"tracing-zipkin-url","onUpdate:modelValue":a[17]||(a[17]=e=>b.validate.meshTracingZipkinURL=e),type:"text",class:"k-input w-100",placeholder:"http://zipkin.url:1234"},null,512),[[n.nr,b.validate.meshTracingZipkinURL]])])),_:1})):(0,n.kq)("",!0)])),_:1})])),metrics:(0,n.w5)((()=>[R,W,(0,n.Wm)(z,{class:"my-6 k-card--small",title:"Metrics Configuration","has-shadow":""},{body:(0,n.w5)((()=>[(0,n.Wm)(x,{title:"Metrics"},{default:(0,n.w5)((()=>[(0,n._)("label",O,[(0,n.wy)((0,n._)("input",{id:"metrics-disabled","onUpdate:modelValue":a[18]||(a[18]=e=>b.validate.metricsEnabled=e),value:"disabled",name:"metrics",type:"radio",class:"k-input mr-2"},null,512),[[n.G2,b.validate.metricsEnabled]]),q]),(0,n._)("label",P,[(0,n.wy)((0,n._)("input",{id:"metrics-enabled","onUpdate:modelValue":a[19]||(a[19]=e=>b.validate.metricsEnabled=e),value:"enabled",name:"metrics",type:"radio",class:"k-input mr-2"},null,512),[[n.G2,b.validate.metricsEnabled]]),j])])),_:1}),"enabled"===b.validate.metricsEnabled?((0,n.wg)(),(0,n.j4)(x,{key:0,title:"Backend name","for-attr":"metrics-name"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"metrics-name","onUpdate:modelValue":a[20]||(a[20]=e=>b.validate.meshMetricsName=e),type:"text",class:"k-input w-100",placeholder:"your-metrics-backend-name"},null,512),[[n.nr,b.validate.meshMetricsName]])])),_:1})):(0,n.kq)("",!0),"enabled"===b.validate.metricsEnabled?((0,n.wg)(),(0,n.j4)(x,{key:1,title:"Type","for-attr":"metrics-type"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("select",{id:"metrics-type","onUpdate:modelValue":a[21]||(a[21]=e=>b.validate.meshMetricsType=e),class:"k-input w-100",name:"metrics-type"},G,512),[[n.bM,b.validate.meshMetricsType]])])),_:1})):(0,n.kq)("",!0),"enabled"===b.validate.metricsEnabled?((0,n.wg)(),(0,n.j4)(x,{key:2,title:"Dataplane port","for-attr":"metrics-dataplane-port"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"metrics-dataplane-port","onUpdate:modelValue":a[22]||(a[22]=e=>b.validate.meshMetricsDataplanePort=e),type:"number",class:"k-input w-100",step:"1",min:"0",max:"65535",placeholder:"1234"},null,512),[[n.nr,b.validate.meshMetricsDataplanePort]])])),_:1})):(0,n.kq)("",!0),"enabled"===b.validate.metricsEnabled?((0,n.wg)(),(0,n.j4)(x,{key:3,title:"Dataplane path","for-attr":"metrics-dataplane-path"},{default:(0,n.w5)((()=>[(0,n.wy)((0,n._)("input",{id:"metrics-dataplane-path","onUpdate:modelValue":a[23]||(a[23]=e=>b.validate.meshMetricsDataplanePath=e),type:"text",class:"k-input w-100"},null,512),[[n.nr,b.validate.meshMetricsDataplanePath]])])),_:1})):(0,n.kq)("",!0)])),_:1})])),complete:(0,n.w5)((()=>[M.codeOutput?((0,n.wg)(),(0,n.iD)("div",F,[!1===b.hideScannerSiblings?((0,n.wg)(),(0,n.iD)("div",K,[Z,(0,n._)("p",null," Since the "+(0,n.zw)(b.productName)+" GUI is read-only mode to follow Ops best practices, please execute the following command in your shell to create the entity. "+(0,n.zw)(b.productName)+" will automatically detect when the new entity has been created. ",1),(0,n.Wm)(re,{loaders:!1,tabs:b.tabs,"initial-tab-override":e.environment,onOnTabChange:M.onTabChange},{kubernetes:(0,n.w5)((()=>[(0,n.Wm)(oe,{title:"Kubernetes","copy-button-text":"Copy Command to Clipboard",lang:"bash",content:M.codeOutput},null,8,["content"])])),universal:(0,n.w5)((()=>[(0,n.Wm)(oe,{title:"Universal","copy-button-text":"Copy Command to Clipboard",lang:"bash",content:M.codeOutput},null,8,["content"])])),_:1},8,["tabs","initial-tab-override","onOnTabChange"])])):(0,n.kq)("",!0),(0,n.Wm)(me,{"loader-function":M.scanForEntity,"should-start":!0,"has-error":b.scanError,"can-complete":b.scanFound,onHideSiblings:M.hideSiblings},{"loading-title":(0,n.w5)((()=>[Y])),"loading-content":(0,n.w5)((()=>[$])),"complete-title":(0,n.w5)((()=>[H])),"complete-content":(0,n.w5)((()=>[(0,n._)("p",null,[J,b.validate.meshName?((0,n.wg)(),(0,n.iD)("strong",Q,(0,n.zw)(b.validate.meshName),1)):(0,n.kq)("",!0),X]),(0,n._)("p",null,[(0,n.Wm)(ce,{appearance:"primary",to:{name:"all-meshes"}},{default:(0,n.w5)((()=>[ee])),_:1})])])),"error-title":(0,n.w5)((()=>[ae])),"error-content":(0,n.w5)((()=>[te])),_:1},8,["loader-function","has-error","can-complete","onHideSiblings"])])):((0,n.wg)(),(0,n.j4)(U,{key:1,appearance:"danger"},{alertMessage:(0,n.w5)((()=>[ne])),_:1}))])),mesh:(0,n.w5)((()=>[le,(0,n._)("p",null," In "+(0,n.zw)(e.title)+", a Mesh resource allows you to define an isolated environment for your data-planes and policies. It's isolated because the mTLS CA you choose can be different from the one configured for our Meshes. Ideally, you will have either a large Mesh with all the workloads, or one Mesh per application for better isolation. ",1),(0,n._)("p",null,[(0,n._)("a",{href:`https://kuma.io/docs/${e.kumaDocsVersion}/policies/mesh/${b.utm}`,target:"_blank"}," Learn More ",8,ie)])])),"did-you-know":(0,n.w5)((()=>[se,de])),_:1},8,["steps","sidebar-content","footer-enabled","next-disabled"])])])}var re=t(33907),ce=t(21180);function me(e,a){return Object.keys(e).filter((e=>!a.includes(e))).map((a=>Object.assign({},{[a]:e[a]}))).reduce(((e,a)=>Object.assign(e,a)),{})}var ue=t(53419),ge=t(88523),pe=t(93897),he=t(34707),be=t(76262),ye=t(71551),ke=t(5129),we=t(27545),fe=t.n(we),_e=t(45689),ve={name:"MeshWizard",components:{FormFragment:pe.Z,TabsWidget:he.Z,StepSkeleton:be.Z,CodeView:ye.Z,EntityScanner:ke.Z},mixins:[ge.Z],data(){return{productName:_e.sG,selectedTab:"",schema:fe(),steps:[{label:"General & Security",slug:"general"},{label:"Logging",slug:"logging"},{label:"Tracing",slug:"tracing"},{label:"Metrics",slug:"metrics"},{label:"Install",slug:"complete"}],tabs:[{hash:"#kubernetes",title:"Kubernetes"},{hash:"#universal",title:"Universal"}],sidebarContent:[{name:"mesh"},{name:"did-you-know"}],formConditions:{mtlsEnabled:!1,loggingEnabled:!1,tracingEnabled:!1,metricsEnabled:!1,loggingType:null},startScanner:!1,scanFound:!1,hideScannerSiblings:!1,scanError:!1,isComplete:!1,validate:{meshName:"",meshCAName:"",meshLoggingBackend:"",meshTracingBackend:"",meshMetricsName:"",meshTracingZipkinURL:"",mtlsEnabled:"disabled",meshCA:"builtin",loggingEnabled:"disabled",loggingType:"tcp",meshLoggingPath:"/",meshLoggingAddress:"127.0.0.1:5000",meshLoggingBackendFormat:"{ start_time: '%START_TIME%', source: '%KUMA_SOURCE_SERVICE%', destination: '%KUMA_DESTINATION_SERVICE%', source_address: '%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%', destination_address: '%UPSTREAM_HOST%', duration_millis: '%DURATION%', bytes_received: '%BYTES_RECEIVED%', bytes_sent: '%BYTES_SENT%' }",tracingEnabled:"disabled",meshTracingType:"zipkin",meshTracingSampling:99.9,metricsEnabled:"disabled",meshMetricsType:"prometheus",meshMetricsDataplanePort:5670,meshMetricsDataplanePath:"/metrics"},vmsg:[],utm:"?utm_source=Kuma&utm_medium=Kuma-GUI"}},computed:{...(0,re.Se)({title:"config/getTagline",kumaDocsVersion:"config/getKumaDocsVersion",environment:"config/getEnvironment"}),codeOutput(){const e=this.schema,a=Object.assign({},e),t=this.validate;if(!t)return;const n="enabled"===t.mtlsEnabled,l="enabled"===t.loggingEnabled,i="enabled"===t.tracingEnabled,s="enabled"===t.metricsEnabled,d={mtls:n,logging:l,tracing:i,metrics:s},o=[];if(Object.entries(d).forEach((e=>{const a=e[1],t=e[0];a?o.filter((e=>e!==t)):o.push(t)})),n){a.mtls.enabled=!0;const e=a.mtls,t=this.validate.meshCA,n=this.validate.meshCAName;e.backends=[],e.enabledBackend=n,e.backends="provided"===t?[{name:n,type:t,conf:{cert:{secret:""},key:{secret:""}}}]:[{name:n,type:t}]}if(l){const e=a.logging.backends[0],n=e.format;e.conf={},e.name=t.meshLoggingBackend,e.type=t.loggingType,e.format=t.meshLoggingBackendFormat||n,"tcp"===t.loggingType?e.conf.address=t.meshLoggingAddress||"127.0.0.1:5000":"file"===t.loggingType&&(e.conf.path=t.meshLoggingPath)}if(i){const e=a.tracing;e.backends[0].conf={},e.defaultBackend=t.meshTracingBackend,e.backends[0].type=t.meshTracingType||"zipkin",e.backends[0].name=t.meshTracingBackend,e.backends[0].conf.sampling=t.meshTracingSampling||100,e.backends[0].conf.url=t.meshTracingZipkinURL}if(s){const e=a.metrics;e.backends[0].conf={},e.enabledBackend=t.meshMetricsName,e.backends[0].type=t.meshMetricsType||"prometheus",e.backends[0].name=t.meshMetricsName,e.backends[0].conf.port=t.meshMetricsDataplanePort||5670,e.backends[0].conf.path=t.meshMetricsDataplanePath||"/metrics"}const r=me(a,o);let c;return c="#kubernetes"===this.selectedTab?{apiVersion:"kuma.io/v1alpha1",kind:"Mesh",metadata:{name:t.meshName},spec:r}:{type:"Mesh",name:t.meshName,...r},this.formatForCLI(c,'" | kumactl apply -f -')},nextDisabled(){const{meshName:e,meshCAName:a,meshLoggingBackend:t,meshTracingBackend:n,meshTracingZipkinURL:l,meshMetricsName:i,mtlsEnabled:s,loggingEnabled:d,tracingEnabled:o,metricsEnabled:r,meshLoggingPath:c,loggingType:m}=this.validate;return!e.length||"enabled"===s&&!a||("1"===this.$route.query.step?"disabled"!==d&&(!t||"file"===m&&!c):"2"===this.$route.query.step?"enabled"===o&&!(n&&l):"3"===this.$route.query.step&&("enabled"===r&&!i))}},watch:{"validate.meshName"(e){const a=(0,ue.GL)(e);this.validate.meshName=a,this.validateMeshName(a)},"validate.meshCAName"(e){this.validate.meshCAName=(0,ue.GL)(e)},"validate.meshLoggingBackend"(e){this.validate.meshLoggingBackend=(0,ue.GL)(e)},"validate.meshTracingBackend"(e){this.validate.meshTracingBackend=(0,ue.GL)(e)},"validate.meshMetricsName"(e){this.validate.meshMetricsName=(0,ue.GL)(e)}},methods:{onTabChange(e){this.selectedTab=e},hideSiblings(){this.hideScannerSiblings=!0},validateMeshName(e){this.vmsg.meshName=e&&""!==e?"":"A Mesh name is required to proceed"},scanForEntity(){const e=this.validate.meshName;this.scanComplete=!1,this.scanError=!1,e&&ce["default"].getMesh({name:e}).then((e=>{e&&e.name.length>0?(this.isRunning=!0,this.scanFound=!0):this.scanError=!0})).catch((e=>{this.scanError=!0,console.error(e)})).finally((()=>{this.scanComplete=!0}))}}},Ee=t(83744);const Te=(0,Ee.Z)(ve,[["render",oe]]);var Se=Te}}]);