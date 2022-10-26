import{C as B,ck as L,cn as V,cS as C,U as D,i as x,o as c,k as N,l as t,a as m,w as s,t as w,b as v,c as g,y as u,m as d,v as h,cT as _,cD as U,j as E}from"./index.563e1198.js";import{_ as I}from"./CodeBlock.vue_vue_type_style_index_0_lang.47deedb5.js";import{F as O}from"./FormatForCLI.72422f2d.js";import{F as R,S as F,E as P}from"./EntityScanner.668ccaa8.js";import{T as K}from"./TabsWidget.7b28b344.js";import"./_commonjsHelpers.f037b798.js";import"./index.58caa11d.js";import"./ErrorBlock.7d4f6d91.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.4521362b.js";function z(o,a){return Object.keys(o).filter(l=>!a.includes(l)).map(l=>Object.assign({},{[l]:o[l]})).reduce((l,T)=>Object.assign(l,T),{})}const j={mtls:{enabledBackend:null,backends:[]},tracing:{defaultBackend:null,backends:[{name:null,type:null}]},logging:{backends:[{name:null,format:'{ "destination": "%KUMA_DESTINATION_SERVICE%", "destinationAddress": "%UPSTREAM_LOCAL_ADDRESS%", "source": "%KUMA_SOURCE_SERVICE%", "sourceAddress": "%KUMA_SOURCE_ADDRESS%", "bytesReceived": "%BYTES_RECEIVED%", "bytesSent": "%BYTES_SENT%"}',type:null}]},metrics:{enabledBackend:null,backends:[{name:null,type:null}]}};function A(){return{meshName:"",meshCAName:"",meshLoggingBackend:"",meshTracingBackend:"",meshMetricsName:"",meshTracingZipkinURL:"",mtlsEnabled:"disabled",meshCA:"builtin",loggingEnabled:"disabled",loggingType:"tcp",meshLoggingPath:"/",meshLoggingAddress:"127.0.0.1:5000",meshLoggingBackendFormat:'{ start_time: "%START_TIME%", source: "%KUMA_SOURCE_SERVICE%", destination: "%KUMA_DESTINATION_SERVICE%", source_address: "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%", destination_address: "%UPSTREAM_HOST%", duration_millis: "%DURATION%", bytes_received: "%BYTES_RECEIVED%", bytes_sent: "%BYTES_SENT%" }',tracingEnabled:"disabled",meshTracingType:"zipkin",meshTracingSampling:99.9,metricsEnabled:"disabled",meshMetricsType:"prometheus",meshMetricsDataplanePort:5670,meshMetricsDataplanePath:"/metrics"}}const Y={name:"MeshWizard",components:{CodeBlock:I,FormFragment:R,TabsWidget:K,StepSkeleton:F,EntityScanner:P},mixins:[O],data(){return{hasStoredMeshData:!1,productName:L,selectedTab:"",schema:j,steps:[{label:"General & Security",slug:"general"},{label:"Logging",slug:"logging"},{label:"Tracing",slug:"tracing"},{label:"Metrics",slug:"metrics"},{label:"Install",slug:"complete"}],tabs:[{hash:"#kubernetes",title:"Kubernetes"},{hash:"#universal",title:"Universal"}],sidebarContent:[{name:"mesh"},{name:"did-you-know"}],formConditions:{mtlsEnabled:!1,loggingEnabled:!1,tracingEnabled:!1,metricsEnabled:!1,loggingType:null},startScanner:!1,scanFound:!1,hideScannerSiblings:!1,scanError:!1,isComplete:!1,validate:A(),vmsg:[],utm:"?utm_source=Kuma&utm_medium=Kuma-GUI"}},computed:{...V({title:"config/getTagline",kumaDocsVersion:"config/getKumaDocsVersion",environment:"config/getEnvironment"}),codeOutput(){const o=this.schema,a=Object.assign({},o),l=this.validate;if(!l)return;const T=l.mtlsEnabled==="enabled",e=l.loggingEnabled==="enabled",p=l.tracingEnabled==="enabled",S=l.metricsEnabled==="enabled",M={mtls:T,logging:e,tracing:p,metrics:S},r=[];if(Object.entries(M).forEach(i=>{const k=i[1],y=i[0];k?r.filter(n=>n!==y):r.push(y)}),T){a.mtls.enabled=!0;const i=a.mtls,k=this.validate.meshCA,y=this.validate.meshCAName;i.backends=[],i.enabledBackend=y,k==="provided"?i.backends=[{name:y,type:k,conf:{cert:{secret:""},key:{secret:""}}}]:i.backends=[{name:y,type:k}]}if(e){const i=a.logging.backends[0],k=i.format;i.conf={},i.name=l.meshLoggingBackend,i.type=l.loggingType,i.format=l.meshLoggingBackendFormat||k,l.loggingType==="tcp"?i.conf.address=l.meshLoggingAddress||"127.0.0.1:5000":l.loggingType==="file"&&(i.conf.path=l.meshLoggingPath)}if(p){const i=a.tracing;i.defaultBackend=l.meshTracingBackend,i.backends[0].type=l.meshTracingType||"zipkin",i.backends[0].name=l.meshTracingBackend,i.backends[0].sampling=l.meshTracingSampling||100,i.backends[0].conf={},i.backends[0].conf.url=l.meshTracingZipkinURL}if(S){const i=a.metrics;i.backends[0].conf={},i.enabledBackend=l.meshMetricsName,i.backends[0].type=l.meshMetricsType||"prometheus",i.backends[0].name=l.meshMetricsName,i.backends[0].conf.port=l.meshMetricsDataplanePort||5670,i.backends[0].conf.path=l.meshMetricsDataplanePath||"/metrics"}const b=z(a,r);let f;return this.selectedTab==="#kubernetes"?(f={apiVersion:"kuma.io/v1alpha1",kind:"Mesh",metadata:{name:l.meshName}},Object.keys(b).length>0&&(f.spec=b)):f={type:"Mesh",name:l.meshName,...b},this.formatForCLI(f,'" | kumactl apply -f -')},nextDisabled(){const{meshName:o,meshCAName:a,meshLoggingBackend:l,meshTracingBackend:T,meshTracingZipkinURL:e,meshMetricsName:p,mtlsEnabled:S,loggingEnabled:M,tracingEnabled:r,metricsEnabled:b,meshLoggingPath:f,loggingType:i}=this.validate;return!o.length||S==="enabled"&&!a?!0:this.$route.query.step==="1"?M==="disabled"?!1:l?i==="file"&&!f:!0:this.$route.query.step==="2"?r==="enabled"&&!(T&&e):this.$route.query.step==="3"?b==="enabled"&&!p:!1}},watch:{"validate.meshName"(o){const a=C(o);this.validate.meshName=a,this.validateMeshName(a)},"validate.meshCAName"(o){this.validate.meshCAName=C(o)},"validate.meshLoggingBackend"(o){this.validate.meshLoggingBackend=C(o)},"validate.meshTracingBackend"(o){this.validate.meshTracingBackend=C(o)},"validate.meshMetricsName"(o){this.validate.meshMetricsName=C(o)}},created(){const o=D.get("createMeshData");o!==null&&(this.validate=o,this.hasStoredMeshData=!0)},methods:{updateStoredData(){D.set("createMeshData",this.validate),this.hasStoredMeshData=!0},resetMeshData(){D.remove("createMeshData"),this.hasStoredMeshData=!1,this.validate=A()},onTabChange(o){this.selectedTab=o},hideSiblings(){this.hideScannerSiblings=!0},validateMeshName(o){!o||o===""?this.vmsg.meshName="A Mesh name is required to proceed":this.vmsg.meshName=""},scanForEntity(){const o=this.validate.meshName;this.scanComplete=!1,this.scanError=!1,o&&x.getMesh({name:o}).then(a=>{a&&a.name.length>0?(this.isRunning=!0,this.scanFound=!0):this.scanError=!0}).catch(a=>{this.scanError=!0,console.error(a)}).finally(()=>{this.scanComplete=!0})}}},W={class:"wizard"},G={class:"wizard__content"},Z=t("code",null,"kubectl",-1),q=t("code",null,"kumactl",-1),H=t("h3",null," To get started, please fill in the following information: ",-1),J={class:"k-input-label mx-2"},Q=t("span",null,"Disabled",-1),X={class:"k-input-label mx-2"},$=t("span",null,"Enabled",-1),ee=t("option",{value:"builtin"}," builtin ",-1),te=t("option",{value:"provided"}," provided ",-1),ae=[ee,te],ne=t("p",{class:"help"}," If you've enabled mTLS, you must select a CA. ",-1),se=t("h3",null," Setup Logging ",-1),le=t("p",null,' You can setup as many logging backends as you need that you can later use to log traffic via the "TrafficLog" policy. In this wizard, we allow you to configure one backend, but you can add more manually if you wish. ',-1),ie={class:"k-input-label mx-2"},oe=t("span",null,"Disabled",-1),re={class:"k-input-label mx-2"},de=t("span",null,"Enabled",-1),ce={key:1},me=t("option",{value:"tcp"}," TCP ",-1),ue=t("option",{value:"file"}," File ",-1),ge=[me,ue],pe=t("h3",null," Setup Tracing ",-1),he=t("p",null,' You can setup as many tracing backends as you need that you can later use to log traffic via the "TrafficTrace" policy. In this wizard we allow you to configure one backend, but you can add more manually as you wish. ',-1),be={class:"k-input-label mx-2"},fe=t("span",null,"Disabled",-1),ke={class:"k-input-label mx-2"},ye=t("span",null,"Enabled",-1),ve=t("option",{value:"zipkin"}," Zipkin ",-1),_e=[ve],Ee=t("h3",null," Setup Metrics ",-1),Te=t("p",null," You can expose metrics from every data-plane on a configurable path and port that a metrics service, like Prometheus, can use to fetch them. ",-1),Se={class:"k-input-label mx-2"},Me=t("span",null,"Disabled",-1),we={class:"k-input-label mx-2"},Ce=t("span",null,"Enabled",-1),Ne=t("option",{value:"prometheus"}," Prometheus ",-1),Ue=[Ne],De={key:0},Ae={key:0},Be=t("h3",null," Install a new Mesh ",-1),Le=t("h3",null,"Searching\u2026",-1),Ve=t("p",null,"We are looking for your mesh.",-1),xe=t("h3",null,"Done!",-1),Ie={key:0},Oe=t("h3",null,"Mesh not found",-1),Re=t("p",null,"We were unable to find your mesh.",-1),Fe=t("p",null," You haven't filled any data out yet! Please return to the first step and fill out your information. ",-1),Pe=t("h3",null,"Mesh",-1),Ke=["href"],ze=t("h3",null,"Did You Know?",-1),je=t("p",null," As you know, the GUI is read-only, but it will be providing instructions to create a new Mesh and verify everything worked well. ",-1);function Ye(o,a,l,T,e,p){const S=E("KButton"),M=E("KAlert"),r=E("FormFragment"),b=E("KCard"),f=E("CodeBlock"),i=E("TabsWidget"),k=E("EntityScanner"),y=E("StepSkeleton");return c(),N("div",W,[t("div",G,[m(y,{steps:e.steps,"sidebar-content":e.sidebarContent,"footer-enabled":e.hideScannerSiblings===!1,"next-disabled":p.nextDisabled,onGoToStep:p.updateStoredData},{general:s(()=>[t("p",null," Welcome to the wizard for creating a new Mesh resource in "+w(e.productName)+". We will be providing you with a few steps that will get you started. ",1),t("p",null,[v(" As you know, the "+w(e.productName)+" GUI is read-only, so at the end of this wizard we will be generating the configuration that you can apply with either ",1),Z,v(" (if you are running in Kubernetes mode) or "),q,v(" / API (if you are running in Universal mode). ")]),H,m(b,{class:"my-6 k-card--small",title:"Mesh Information","has-shadow":""},{body:s(()=>[e.hasStoredMeshData?(c(),g(M,{key:0,class:"reset-mesh-data-alert",appearance:"info"},{alertMessage:s(()=>[v(" Want to start with an empty slate? ")]),actionButtons:s(()=>[m(S,{apperance:"outline",onClick:p.resetMeshData},{default:s(()=>[v(" Reset to defaults ")]),_:1},8,["onClick"])]),_:1})):u("",!0),m(r,{class:"mt-4",title:"Mesh name","for-attr":"mesh-name"},{default:s(()=>[d(t("input",{id:"mesh-name","onUpdate:modelValue":a[0]||(a[0]=n=>e.validate.meshName=n),type:"text",class:"k-input w-100",placeholder:"your-mesh-name",required:""},null,512),[[h,e.validate.meshName]]),e.vmsg.meshName?(c(),g(M,{key:0,appearance:"danger",size:"small","alert-message":e.vmsg.meshName},null,8,["alert-message"])):u("",!0)]),_:1}),m(r,{class:"mt-4",title:"Mutual TLS"},{default:s(()=>[t("label",J,[d(t("input",{ref:"mtlsDisabled","onUpdate:modelValue":a[1]||(a[1]=n=>e.validate.mtlsEnabled=n),value:"disabled",name:"mtls",type:"radio",class:"k-input mr-2"},null,512),[[_,e.validate.mtlsEnabled]]),Q]),t("label",X,[d(t("input",{id:"mtls-enabled","onUpdate:modelValue":a[2]||(a[2]=n=>e.validate.mtlsEnabled=n),value:"enabled",name:"mtls",type:"radio",class:"k-input mr-2"},null,512),[[_,e.validate.mtlsEnabled]]),$])]),_:1}),e.validate.mtlsEnabled==="enabled"?(c(),g(r,{key:1,class:"mt-4",title:"Certificate name","for-attr":"certificate-name"},{default:s(()=>[d(t("input",{id:"certificate-name","onUpdate:modelValue":a[3]||(a[3]=n=>e.validate.meshCAName=n),type:"text",class:"k-input w-100",placeholder:"your-certificate-name"},null,512),[[h,e.validate.meshCAName]])]),_:1})):u("",!0),e.validate.mtlsEnabled==="enabled"?(c(),g(r,{key:2,class:"mt-4",title:"Certificate Authority","for-attr":"certificate-authority"},{default:s(()=>[d(t("select",{id:"certificate-authority","onUpdate:modelValue":a[4]||(a[4]=n=>e.validate.meshCA=n),class:"k-input w-100",name:"certificate-authority"},ae,512),[[U,e.validate.meshCA]]),ne]),_:1})):u("",!0)]),_:1})]),logging:s(()=>[se,le,m(b,{class:"my-6 k-card--small",title:"Logging Configuration","has-shadow":""},{body:s(()=>[m(r,{title:"Logging"},{default:s(()=>[t("label",ie,[d(t("input",{id:"logging-disabled","onUpdate:modelValue":a[5]||(a[5]=n=>e.validate.loggingEnabled=n),value:"disabled",name:"logging",type:"radio",class:"k-input mr-2"},null,512),[[_,e.validate.loggingEnabled]]),oe]),t("label",re,[d(t("input",{id:"logging-enabled","onUpdate:modelValue":a[6]||(a[6]=n=>e.validate.loggingEnabled=n),value:"enabled",name:"logging",type:"radio",class:"k-input mr-2"},null,512),[[_,e.validate.loggingEnabled]]),de])]),_:1}),e.validate.loggingEnabled==="enabled"?(c(),g(r,{key:0,class:"mt-4",title:"Backend name","for-attr":"backend-name"},{default:s(()=>[d(t("input",{id:"backend-name","onUpdate:modelValue":a[7]||(a[7]=n=>e.validate.meshLoggingBackend=n),type:"text",class:"k-input w-100",placeholder:"your-backend-name"},null,512),[[h,e.validate.meshLoggingBackend]])]),_:1})):u("",!0),e.validate.loggingEnabled==="enabled"?(c(),N("div",ce,[m(r,{class:"mt-4",title:"Type"},{default:s(()=>[d(t("select",{id:"logging-type",ref:"loggingTypeSelect","onUpdate:modelValue":a[8]||(a[8]=n=>e.validate.loggingType=n),class:"k-input w-100",name:"logging-type"},ge,512),[[U,e.validate.loggingType]])]),_:1}),e.validate.loggingType==="file"?(c(),g(r,{key:0,class:"mt-4",title:"Path","for-attr":"backend-address"},{default:s(()=>[d(t("input",{id:"backend-address","onUpdate:modelValue":a[9]||(a[9]=n=>e.validate.meshLoggingPath=n),type:"text",class:"k-input w-100"},null,512),[[h,e.validate.meshLoggingPath]])]),_:1})):u("",!0),e.validate.loggingType==="tcp"?(c(),g(r,{key:1,class:"mt-4",title:"Address","for-attr":"backend-address"},{default:s(()=>[d(t("input",{id:"backend-address","onUpdate:modelValue":a[10]||(a[10]=n=>e.validate.meshLoggingAddress=n),type:"text",class:"k-input w-100"},null,512),[[h,e.validate.meshLoggingAddress]])]),_:1})):u("",!0),m(r,{class:"mt-4",title:"Format","for-attr":"backend-format"},{default:s(()=>[d(t("textarea",{id:"backend-format","onUpdate:modelValue":a[11]||(a[11]=n=>e.validate.meshLoggingBackendFormat=n),class:"k-input w-100 code-sample",rows:"12"},null,512),[[h,e.validate.meshLoggingBackendFormat]])]),_:1})])):u("",!0)]),_:1})]),tracing:s(()=>[pe,he,m(b,{class:"my-6 k-card--small",title:"Tracing Configuration","has-shadow":""},{body:s(()=>[m(r,{title:"Tracing"},{default:s(()=>[t("label",be,[d(t("input",{id:"tracing-disabled","onUpdate:modelValue":a[12]||(a[12]=n=>e.validate.tracingEnabled=n),value:"disabled",name:"tracing",type:"radio",class:"k-input mr-2"},null,512),[[_,e.validate.tracingEnabled]]),fe]),t("label",ke,[d(t("input",{id:"tracing-enabled","onUpdate:modelValue":a[13]||(a[13]=n=>e.validate.tracingEnabled=n),value:"enabled",name:"tracing",type:"radio",class:"k-input mr-2"},null,512),[[_,e.validate.tracingEnabled]]),ye])]),_:1}),e.validate.tracingEnabled==="enabled"?(c(),g(r,{key:0,class:"mt-4",title:"Backend name","for-attr":"tracing-backend-name"},{default:s(()=>[d(t("input",{id:"tracing-backend-name","onUpdate:modelValue":a[14]||(a[14]=n=>e.validate.meshTracingBackend=n),type:"text",class:"k-input w-100",placeholder:"your-tracing-backend-name"},null,512),[[h,e.validate.meshTracingBackend]])]),_:1})):u("",!0),e.validate.tracingEnabled==="enabled"?(c(),g(r,{key:1,class:"mt-4",title:"Type","for-attr":"tracing-type"},{default:s(()=>[d(t("select",{id:"tracing-type","onUpdate:modelValue":a[15]||(a[15]=n=>e.validate.meshTracingType=n),class:"k-input w-100",name:"tracing-type"},_e,512),[[U,e.validate.meshTracingType]])]),_:1})):u("",!0),e.validate.tracingEnabled==="enabled"?(c(),g(r,{key:2,class:"mt-4",title:"Sampling","for-attr":"tracing-sampling"},{default:s(()=>[d(t("input",{id:"tracing-sampling","onUpdate:modelValue":a[16]||(a[16]=n=>e.validate.meshTracingSampling=n),type:"number",class:"k-input w-100",step:"0.1",min:"0",max:"100"},null,512),[[h,e.validate.meshTracingSampling]])]),_:1})):u("",!0),e.validate.tracingEnabled==="enabled"?(c(),g(r,{key:3,class:"mt-4",title:"URL","for-attr":"tracing-zipkin-url"},{default:s(()=>[d(t("input",{id:"tracing-zipkin-url","onUpdate:modelValue":a[17]||(a[17]=n=>e.validate.meshTracingZipkinURL=n),type:"text",class:"k-input w-100",placeholder:"http://zipkin.url:1234"},null,512),[[h,e.validate.meshTracingZipkinURL]])]),_:1})):u("",!0)]),_:1})]),metrics:s(()=>[Ee,Te,m(b,{class:"my-6 k-card--small",title:"Metrics Configuration","has-shadow":""},{body:s(()=>[m(r,{title:"Metrics"},{default:s(()=>[t("label",Se,[d(t("input",{id:"metrics-disabled","onUpdate:modelValue":a[18]||(a[18]=n=>e.validate.metricsEnabled=n),value:"disabled",name:"metrics",type:"radio",class:"k-input mr-2"},null,512),[[_,e.validate.metricsEnabled]]),Me]),t("label",we,[d(t("input",{id:"metrics-enabled","onUpdate:modelValue":a[19]||(a[19]=n=>e.validate.metricsEnabled=n),value:"enabled",name:"metrics",type:"radio",class:"k-input mr-2"},null,512),[[_,e.validate.metricsEnabled]]),Ce])]),_:1}),e.validate.metricsEnabled==="enabled"?(c(),g(r,{key:0,class:"mt-4",title:"Backend name","for-attr":"metrics-name"},{default:s(()=>[d(t("input",{id:"metrics-name","onUpdate:modelValue":a[20]||(a[20]=n=>e.validate.meshMetricsName=n),type:"text",class:"k-input w-100",placeholder:"your-metrics-backend-name"},null,512),[[h,e.validate.meshMetricsName]])]),_:1})):u("",!0),e.validate.metricsEnabled==="enabled"?(c(),g(r,{key:1,class:"mt-4",title:"Type","for-attr":"metrics-type"},{default:s(()=>[d(t("select",{id:"metrics-type","onUpdate:modelValue":a[21]||(a[21]=n=>e.validate.meshMetricsType=n),class:"k-input w-100",name:"metrics-type"},Ue,512),[[U,e.validate.meshMetricsType]])]),_:1})):u("",!0),e.validate.metricsEnabled==="enabled"?(c(),g(r,{key:2,class:"mt-4",title:"Dataplane port","for-attr":"metrics-dataplane-port"},{default:s(()=>[d(t("input",{id:"metrics-dataplane-port","onUpdate:modelValue":a[22]||(a[22]=n=>e.validate.meshMetricsDataplanePort=n),type:"number",class:"k-input w-100",step:"1",min:"0",max:"65535",placeholder:"1234"},null,512),[[h,e.validate.meshMetricsDataplanePort]])]),_:1})):u("",!0),e.validate.metricsEnabled==="enabled"?(c(),g(r,{key:3,class:"mt-4",title:"Dataplane path","for-attr":"metrics-dataplane-path"},{default:s(()=>[d(t("input",{id:"metrics-dataplane-path","onUpdate:modelValue":a[23]||(a[23]=n=>e.validate.meshMetricsDataplanePath=n),type:"text",class:"k-input w-100"},null,512),[[h,e.validate.meshMetricsDataplanePath]])]),_:1})):u("",!0)]),_:1})]),complete:s(()=>[p.codeOutput?(c(),N("div",De,[e.hideScannerSiblings===!1?(c(),N("div",Ae,[Be,t("p",null," Since the "+w(e.productName)+" GUI is read-only mode to follow Ops best practices, please execute the following command in your shell to create the entity. "+w(e.productName)+" will automatically detect when the new entity has been created. ",1),m(i,{tabs:e.tabs,"initial-tab-override":o.environment,onOnTabChange:p.onTabChange},{kubernetes:s(()=>[m(f,{id:"code-block-kubernetes-command",language:"bash",code:p.codeOutput},null,8,["code"])]),universal:s(()=>[m(f,{id:"code-block-universal-command",language:"bash",code:p.codeOutput},null,8,["code"])]),_:1},8,["tabs","initial-tab-override","onOnTabChange"])])):u("",!0),m(k,{"loader-function":p.scanForEntity,"should-start":!0,"has-error":e.scanError,"can-complete":e.scanFound,onHideSiblings:p.hideSiblings},{"loading-title":s(()=>[Le]),"loading-content":s(()=>[Ve]),"complete-title":s(()=>[xe]),"complete-content":s(()=>[t("p",null,[v(" Your mesh "),e.validate.meshName?(c(),N("strong",Ie,w(e.validate.meshName),1)):u("",!0),v(" was found! ")]),t("p",null,[m(S,{appearance:"primary",to:{name:"mesh-detail-view",params:{mesh:e.validate.meshName}}},{default:s(()=>[v(" Go to mesh "+w(e.validate.meshName),1)]),_:1},8,["to"])])]),"error-title":s(()=>[Oe]),"error-content":s(()=>[Re]),_:1},8,["loader-function","has-error","can-complete","onHideSiblings"])])):(c(),g(M,{key:1,appearance:"danger"},{alertMessage:s(()=>[Fe]),_:1}))]),mesh:s(()=>[Pe,t("p",null," In "+w(o.title)+", a Mesh resource allows you to define an isolated environment for your data-planes and policies. It's isolated because the mTLS CA you choose can be different from the one configured for our Meshes. Ideally, you will have either a large Mesh with all the workloads, or one Mesh per application for better isolation. ",1),t("p",null,[t("a",{href:`https://kuma.io/docs/${o.kumaDocsVersion}/policies/mesh/${e.utm}`,target:"_blank"}," Learn More ",8,Ke)])]),"did-you-know":s(()=>[ze,je]),_:1},8,["steps","sidebar-content","footer-enabled","next-disabled","onGoToStep"])])])}const et=B(Y,[["render",Ye]]);export{et as default};
