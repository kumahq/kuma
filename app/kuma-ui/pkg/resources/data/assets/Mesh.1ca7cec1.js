import{E as V,z as I,P as R,L as x,ck as O,cn as P,cR as N,R as A,k as K,i as v,o as u,j as D,l as t,a as g,w as l,t as C,e as n,c as h,A as p,m,v as f,cS as E,cC as U,C as F,D as z}from"./index.c8e7c817.js";import{_ as j}from"./CodeBlock.vue_vue_type_style_index_0_lang.660b597c.js";import{f as Y}from"./formatForCLI.199be697.js";import{F as W,S as G,E as Z}from"./EntityScanner.44705ff2.js";import{T as q}from"./TabsWidget.cef20a04.js";import"./_commonjsHelpers.f037b798.js";import"./index.58caa11d.js";import"./ErrorBlock.26868ad8.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.5817f994.js";const H={mtls:{enabledBackend:null,backends:[]},tracing:{defaultBackend:null,backends:[{name:null,type:null}]},logging:{backends:[{name:null,format:'{ "destination": "%KUMA_DESTINATION_SERVICE%", "destinationAddress": "%UPSTREAM_LOCAL_ADDRESS%", "source": "%KUMA_SOURCE_SERVICE%", "sourceAddress": "%KUMA_SOURCE_ADDRESS%", "bytesReceived": "%BYTES_RECEIVED%", "bytesSent": "%BYTES_SENT%"}',type:null}]},metrics:{enabledBackend:null,backends:[{name:null,type:null}]}};function B(){return{meshName:"",meshCAName:"",meshLoggingBackend:"",meshTracingBackend:"",meshMetricsName:"",meshTracingZipkinURL:"",mtlsEnabled:"disabled",meshCA:"builtin",loggingEnabled:"disabled",loggingType:"tcp",meshLoggingPath:"/",meshLoggingAddress:"127.0.0.1:5000",meshLoggingBackendFormat:'{ start_time: "%START_TIME%", source: "%KUMA_SOURCE_SERVICE%", destination: "%KUMA_DESTINATION_SERVICE%", source_address: "%KUMA_SOURCE_ADDRESS_WITHOUT_PORT%", destination_address: "%UPSTREAM_HOST%", duration_millis: "%DURATION%", bytes_received: "%BYTES_RECEIVED%", bytes_sent: "%BYTES_SENT%" }',tracingEnabled:"disabled",meshTracingType:"zipkin",meshTracingSampling:99.9,metricsEnabled:"disabled",meshMetricsType:"prometheus",meshMetricsDataplanePort:5670,meshMetricsDataplanePath:"/metrics"}}function J(i,a){return Object.keys(i).filter(o=>!a.includes(o)).map(o=>Object.assign({},{[o]:i[o]})).reduce((o,T)=>Object.assign(o,T),{})}const Q={name:"MeshWizard",components:{CodeBlock:j,FormFragment:W,TabsWidget:q,StepSkeleton:G,EntityScanner:Z,KAlert:I,KButton:R,KCard:x},data(){return{hasStoredMeshData:!1,productName:O,selectedTab:"",schema:H,steps:[{label:"General & Security",slug:"general"},{label:"Logging",slug:"logging"},{label:"Tracing",slug:"tracing"},{label:"Metrics",slug:"metrics"},{label:"Install",slug:"complete"}],tabs:[{hash:"#kubernetes",title:"Kubernetes"},{hash:"#universal",title:"Universal"}],sidebarContent:[{name:"mesh"},{name:"did-you-know"}],formConditions:{mtlsEnabled:!1,loggingEnabled:!1,tracingEnabled:!1,metricsEnabled:!1,loggingType:null},startScanner:!1,scanFound:!1,hideScannerSiblings:!1,scanError:!1,isComplete:!1,validate:B(),vmsg:[],utm:"?utm_source=Kuma&utm_medium=Kuma-GUI"}},computed:{...P({title:"config/getTagline",kumaDocsVersion:"config/getKumaDocsVersion",environment:"config/getEnvironment"}),codeOutput(){const i=this.schema,a=Object.assign({},i),o=this.validate;if(!o)return;const T=o.mtlsEnabled==="enabled",e=o.loggingEnabled==="enabled",b=o.tracingEnabled==="enabled",S=o.metricsEnabled==="enabled",M={mtls:T,logging:e,tracing:b,metrics:S},c=[];if(Object.entries(M).forEach(r=>{const _=r[1],s=r[0];_?c.filter(L=>L!==s):c.push(s)}),T){a.mtls.enabled=!0;const r=a.mtls,_=this.validate.meshCA,s=this.validate.meshCAName;r.backends=[],r.enabledBackend=s,_==="provided"?r.backends=[{name:s,type:_,conf:{cert:{secret:""},key:{secret:""}}}]:r.backends=[{name:s,type:_}]}if(e){const r=a.logging.backends[0],_=r.format;r.conf={},r.name=o.meshLoggingBackend,r.type=o.loggingType,r.format=o.meshLoggingBackendFormat||_,o.loggingType==="tcp"?r.conf.address=o.meshLoggingAddress||"127.0.0.1:5000":o.loggingType==="file"&&(r.conf.path=o.meshLoggingPath)}if(b){const r=a.tracing;r.defaultBackend=o.meshTracingBackend,r.backends[0].type=o.meshTracingType||"zipkin",r.backends[0].name=o.meshTracingBackend,r.backends[0].sampling=o.meshTracingSampling||100,r.backends[0].conf={},r.backends[0].conf.url=o.meshTracingZipkinURL}if(S){const r=a.metrics;r.backends[0].conf={},r.enabledBackend=o.meshMetricsName,r.backends[0].type=o.meshMetricsType||"prometheus",r.backends[0].name=o.meshMetricsName,r.backends[0].conf.port=o.meshMetricsDataplanePort||5670,r.backends[0].conf.path=o.meshMetricsDataplanePath||"/metrics"}const k=J(a,c);let y,w;return this.selectedTab==="#kubernetes"?(w="kubectl",y={apiVersion:"kuma.io/v1alpha1",kind:"Mesh",metadata:{name:o.meshName}},Object.keys(k).length>0&&(y.spec=k)):(w="kumactl",y={type:"Mesh",name:o.meshName,...k}),Y(y,`" | ${w} apply -f -`)},nextDisabled(){const{meshName:i,meshCAName:a,meshLoggingBackend:o,meshTracingBackend:T,meshTracingZipkinURL:e,meshMetricsName:b,mtlsEnabled:S,loggingEnabled:M,tracingEnabled:c,metricsEnabled:k,meshLoggingPath:y,loggingType:w}=this.validate;return!i.length||S==="enabled"&&!a?!0:this.$route.query.step==="1"?M==="disabled"?!1:o?w==="file"&&!y:!0:this.$route.query.step==="2"?c==="enabled"&&!(T&&e):this.$route.query.step==="3"?k==="enabled"&&!b:!1}},watch:{"validate.meshName"(i){const a=N(i);this.validate.meshName=a,this.validateMeshName(a)},"validate.meshCAName"(i){this.validate.meshCAName=N(i)},"validate.meshLoggingBackend"(i){this.validate.meshLoggingBackend=N(i)},"validate.meshTracingBackend"(i){this.validate.meshTracingBackend=N(i)},"validate.meshMetricsName"(i){this.validate.meshMetricsName=N(i)}},created(){const i=A.get("createMeshData");i!==null&&(this.validate=i,this.hasStoredMeshData=!0)},methods:{updateStoredData(){A.set("createMeshData",this.validate),this.hasStoredMeshData=!0},resetMeshData(){A.remove("createMeshData"),this.hasStoredMeshData=!1,this.validate=B()},onTabChange(i){this.selectedTab=i},hideSiblings(){this.hideScannerSiblings=!0},validateMeshName(i){!i||i===""?this.vmsg.meshName="A Mesh name is required to proceed":this.vmsg.meshName=""},scanForEntity(){const i=this.validate.meshName;this.scanComplete=!1,this.scanError=!1,i&&K.getMesh({name:i}).then(a=>{a&&a.name.length>0?(this.isRunning=!0,this.scanFound=!0):this.scanError=!0}).catch(a=>{this.scanError=!0,console.error(a)}).finally(()=>{this.scanComplete=!0})}}},d=i=>(F("data-v-7dc40dea"),i=i(),z(),i),X={class:"wizard"},$={class:"wizard__content"},ee=d(()=>t("code",null,"kubectl",-1)),te=d(()=>t("code",null,"kumactl",-1)),ae=d(()=>t("h3",null,`
            To get started, please fill in the following information:
          `,-1)),ne={class:"k-input-label mx-2"},se=d(()=>t("span",null,"Disabled",-1)),le={class:"k-input-label mx-2"},ie=d(()=>t("span",null,"Enabled",-1)),oe=d(()=>t("option",{value:"builtin"},`
                    builtin
                  `,-1)),de=d(()=>t("option",{value:"provided"},`
                    provided
                  `,-1)),re=d(()=>t("p",{class:"help"},`
                  If you've enabled mTLS, you must select a CA.
                `,-1)),ce=d(()=>t("h3",null,`
            Setup Logging
          `,-1)),me=d(()=>t("p",null,`
            You can setup as many logging backends as you need that you can later
            use to log traffic via the "TrafficLog" policy. In this wizard,
            we allow you to configure one backend, but you can add more manually
            if you wish.
          `,-1)),ue={class:"k-input-label mx-2"},ge=d(()=>t("span",null,"Disabled",-1)),pe={class:"k-input-label mx-2"},he=d(()=>t("span",null,"Enabled",-1)),be={key:1},fe=d(()=>t("option",{value:"tcp"},`
                      TCP
                    `,-1)),ke=d(()=>t("option",{value:"file"},`
                      File
                    `,-1)),ye=d(()=>t("h3",null,`
            Setup Tracing
          `,-1)),_e=d(()=>t("p",null,`
            You can setup as many tracing backends as you need that you can later
            use to log traffic via the "TrafficTrace" policy. In this
            wizard we allow you to configure one backend, but you can add more
            manually as you wish.
          `,-1)),ve={class:"k-input-label mx-2"},Ee=d(()=>t("span",null,"Disabled",-1)),Te={class:"k-input-label mx-2"},Se=d(()=>t("span",null,"Enabled",-1)),Me=d(()=>t("option",{value:"zipkin"},`
                    Zipkin
                  `,-1)),we=[Me],Ce=d(()=>t("h3",null,`
            Setup Metrics
          `,-1)),Ne=d(()=>t("p",null,`
            You can expose metrics from every data-plane on a configurable path
            and port that a metrics service, like Prometheus, can use to fetch them.
          `,-1)),De={class:"k-input-label mx-2"},Ue=d(()=>t("span",null,"Disabled",-1)),Ae={class:"k-input-label mx-2"},Be=d(()=>t("span",null,"Enabled",-1)),Le=d(()=>t("option",{value:"prometheus"},`
                    Prometheus
                  `,-1)),Ve=[Le],Ie={key:0},Re={key:0},xe=d(()=>t("h3",null,`
                Install a new Mesh
              `,-1)),Oe=d(()=>t("h3",null,"Searching\u2026",-1)),Pe=d(()=>t("p",null,"We are looking for your mesh.",-1)),Ke=d(()=>t("h3",null,"Done!",-1)),Fe={key:0},ze=d(()=>t("h3",null,"Mesh not found",-1)),je=d(()=>t("p",null,"We were unable to find your mesh.",-1)),Ye=d(()=>t("p",null,`
                You haven't filled any data out yet! Please return to the first
                step and fill out your information.
              `,-1)),We=d(()=>t("h3",null,"Mesh",-1)),Ge=["href"],Ze=d(()=>t("h3",null,"Did You Know?",-1)),qe=d(()=>t("p",null,`
            As you know, the GUI is read-only, but it will be providing instructions
            to create a new Mesh and verify everything worked well.
          `,-1));function He(i,a,o,T,e,b){const S=v("KButton"),M=v("KAlert"),c=v("FormFragment"),k=v("KCard"),y=v("CodeBlock"),w=v("TabsWidget"),r=v("EntityScanner"),_=v("StepSkeleton");return u(),D("div",X,[t("div",$,[g(_,{steps:e.steps,"sidebar-content":e.sidebarContent,"footer-enabled":e.hideScannerSiblings===!1,"next-disabled":b.nextDisabled,onGoToStep:b.updateStoredData},{general:l(()=>[t("p",null,`
            Welcome to the wizard for creating a new Mesh resource in `+C(e.productName)+`.
            We will be providing you with a few steps that will get you started.
          `,1),n(),t("p",null,[n(`
            As you know, the `+C(e.productName)+` GUI is read-only, so at the end of this wizard
            we will be generating the configuration that you can apply with either
            `,1),ee,n(` (if you are running in Kubernetes mode) or
            `),te,n(` / API (if you are running in Universal mode).
          `)]),n(),ae,n(),g(k,{class:"my-6",title:"Mesh Information","has-shadow":""},{body:l(()=>[e.hasStoredMeshData?(u(),h(M,{key:0,class:"reset-mesh-data-alert",appearance:"info"},{alertMessage:l(()=>[n(`
                  Want to start with an empty slate?
                `)]),actionButtons:l(()=>[g(S,{apperance:"outline",onClick:b.resetMeshData},{default:l(()=>[n(`
                    Reset to defaults
                  `)]),_:1},8,["onClick"])]),_:1})):p("",!0),n(),g(c,{class:"mt-4",title:"Mesh name","for-attr":"mesh-name"},{default:l(()=>[m(t("input",{id:"mesh-name","onUpdate:modelValue":a[0]||(a[0]=s=>e.validate.meshName=s),type:"text",class:"k-input w-100","data-testid":"mesh-name",placeholder:"your-mesh-name",required:""},null,512),[[f,e.validate.meshName]]),n(),e.vmsg.meshName?(u(),h(M,{key:0,appearance:"danger",size:"small","alert-message":e.vmsg.meshName},null,8,["alert-message"])):p("",!0)]),_:1}),n(),g(c,{class:"mt-4",title:"Mutual TLS"},{default:l(()=>[t("label",ne,[m(t("input",{ref:"mtlsDisabled","onUpdate:modelValue":a[1]||(a[1]=s=>e.validate.mtlsEnabled=s),value:"disabled",name:"mtls",type:"radio",class:"k-input mr-2","data-testid":"mesh-mtls-disabled"},null,512),[[E,e.validate.mtlsEnabled]]),n(),se]),n(),t("label",le,[m(t("input",{id:"mtls-enabled","onUpdate:modelValue":a[2]||(a[2]=s=>e.validate.mtlsEnabled=s),value:"enabled",name:"mtls",type:"radio",class:"k-input mr-2","data-testid":"mesh-mtls-enabled"},null,512),[[E,e.validate.mtlsEnabled]]),n(),ie])]),_:1}),n(),e.validate.mtlsEnabled==="enabled"?(u(),h(c,{key:1,class:"mt-4",title:"Certificate name","for-attr":"certificate-name"},{default:l(()=>[m(t("input",{id:"certificate-name","onUpdate:modelValue":a[3]||(a[3]=s=>e.validate.meshCAName=s),type:"text",class:"k-input w-100",placeholder:"your-certificate-name","data-testid":"mesh-certificate-name"},null,512),[[f,e.validate.meshCAName]])]),_:1})):p("",!0),n(),e.validate.mtlsEnabled==="enabled"?(u(),h(c,{key:2,class:"mt-4",title:"Certificate Authority","for-attr":"certificate-authority"},{default:l(()=>[m(t("select",{id:"certificate-authority","onUpdate:modelValue":a[4]||(a[4]=s=>e.validate.meshCA=s),class:"k-input w-100",name:"certificate-authority"},[oe,n(),de],512),[[U,e.validate.meshCA]]),n(),re]),_:1})):p("",!0)]),_:1})]),logging:l(()=>[ce,n(),me,n(),g(k,{class:"my-6",title:"Logging Configuration","has-shadow":""},{body:l(()=>[g(c,{title:"Logging"},{default:l(()=>[t("label",ue,[m(t("input",{id:"logging-disabled","onUpdate:modelValue":a[5]||(a[5]=s=>e.validate.loggingEnabled=s),value:"disabled",name:"logging",type:"radio",class:"k-input mr-2","data-testid":"mesh-logging-disabled"},null,512),[[E,e.validate.loggingEnabled]]),n(),ge]),n(),t("label",pe,[m(t("input",{id:"logging-enabled","onUpdate:modelValue":a[6]||(a[6]=s=>e.validate.loggingEnabled=s),value:"enabled",name:"logging",type:"radio",class:"k-input mr-2","data-testid":"mesh-logging-enabled"},null,512),[[E,e.validate.loggingEnabled]]),n(),he])]),_:1}),n(),e.validate.loggingEnabled==="enabled"?(u(),h(c,{key:0,class:"mt-4",title:"Backend name","for-attr":"backend-name"},{default:l(()=>[m(t("input",{id:"backend-name","onUpdate:modelValue":a[7]||(a[7]=s=>e.validate.meshLoggingBackend=s),type:"text",class:"k-input w-100",placeholder:"your-backend-name","data-testid":"mesh-logging-backend-name"},null,512),[[f,e.validate.meshLoggingBackend]])]),_:1})):p("",!0),n(),e.validate.loggingEnabled==="enabled"?(u(),D("div",be,[g(c,{class:"mt-4",title:"Type"},{default:l(()=>[m(t("select",{id:"logging-type",ref:"loggingTypeSelect","onUpdate:modelValue":a[8]||(a[8]=s=>e.validate.loggingType=s),class:"k-input w-100",name:"logging-type"},[fe,n(),ke],512),[[U,e.validate.loggingType]])]),_:1}),n(),e.validate.loggingType==="file"?(u(),h(c,{key:0,class:"mt-4",title:"Path","for-attr":"backend-address"},{default:l(()=>[m(t("input",{id:"backend-address","onUpdate:modelValue":a[9]||(a[9]=s=>e.validate.meshLoggingPath=s),type:"text",class:"k-input w-100"},null,512),[[f,e.validate.meshLoggingPath]])]),_:1})):p("",!0),n(),e.validate.loggingType==="tcp"?(u(),h(c,{key:1,class:"mt-4",title:"Address","for-attr":"backend-address"},{default:l(()=>[m(t("input",{id:"backend-address","onUpdate:modelValue":a[10]||(a[10]=s=>e.validate.meshLoggingAddress=s),type:"text",class:"k-input w-100"},null,512),[[f,e.validate.meshLoggingAddress]])]),_:1})):p("",!0),n(),g(c,{class:"mt-4",title:"Format","for-attr":"backend-format"},{default:l(()=>[m(t("textarea",{id:"backend-format","onUpdate:modelValue":a[11]||(a[11]=s=>e.validate.meshLoggingBackendFormat=s),class:"k-input w-100 code-sample",rows:"12"},null,512),[[f,e.validate.meshLoggingBackendFormat]])]),_:1})])):p("",!0)]),_:1})]),tracing:l(()=>[ye,n(),_e,n(),g(k,{class:"my-6",title:"Tracing Configuration","has-shadow":""},{body:l(()=>[g(c,{title:"Tracing"},{default:l(()=>[t("label",ve,[m(t("input",{id:"tracing-disabled","onUpdate:modelValue":a[12]||(a[12]=s=>e.validate.tracingEnabled=s),value:"disabled",name:"tracing",type:"radio",class:"k-input mr-2"},null,512),[[E,e.validate.tracingEnabled]]),n(),Ee]),n(),t("label",Te,[m(t("input",{id:"tracing-enabled","onUpdate:modelValue":a[13]||(a[13]=s=>e.validate.tracingEnabled=s),value:"enabled",name:"tracing",type:"radio",class:"k-input mr-2","data-testid":"mesh-tracing-enabled"},null,512),[[E,e.validate.tracingEnabled]]),n(),Se])]),_:1}),n(),e.validate.tracingEnabled==="enabled"?(u(),h(c,{key:0,class:"mt-4",title:"Backend name","for-attr":"tracing-backend-name"},{default:l(()=>[m(t("input",{id:"tracing-backend-name","onUpdate:modelValue":a[14]||(a[14]=s=>e.validate.meshTracingBackend=s),type:"text",class:"k-input w-100",placeholder:"your-tracing-backend-name","data-testid":"mesh-tracing-backend-name"},null,512),[[f,e.validate.meshTracingBackend]])]),_:1})):p("",!0),n(),e.validate.tracingEnabled==="enabled"?(u(),h(c,{key:1,class:"mt-4",title:"Type","for-attr":"tracing-type"},{default:l(()=>[m(t("select",{id:"tracing-type","onUpdate:modelValue":a[15]||(a[15]=s=>e.validate.meshTracingType=s),class:"k-input w-100",name:"tracing-type"},we,512),[[U,e.validate.meshTracingType]])]),_:1})):p("",!0),n(),e.validate.tracingEnabled==="enabled"?(u(),h(c,{key:2,class:"mt-4",title:"Sampling","for-attr":"tracing-sampling"},{default:l(()=>[m(t("input",{id:"tracing-sampling","onUpdate:modelValue":a[16]||(a[16]=s=>e.validate.meshTracingSampling=s),type:"number",class:"k-input w-100",step:"0.1",min:"0",max:"100"},null,512),[[f,e.validate.meshTracingSampling]])]),_:1})):p("",!0),n(),e.validate.tracingEnabled==="enabled"?(u(),h(c,{key:3,class:"mt-4",title:"URL","for-attr":"tracing-zipkin-url"},{default:l(()=>[m(t("input",{id:"tracing-zipkin-url","onUpdate:modelValue":a[17]||(a[17]=s=>e.validate.meshTracingZipkinURL=s),type:"text",class:"k-input w-100",placeholder:"http://zipkin.url:1234","data-testid":"mesh-tracing-url"},null,512),[[f,e.validate.meshTracingZipkinURL]])]),_:1})):p("",!0)]),_:1})]),metrics:l(()=>[Ce,n(),Ne,n(),g(k,{class:"my-6",title:"Metrics Configuration","has-shadow":""},{body:l(()=>[g(c,{title:"Metrics"},{default:l(()=>[t("label",De,[m(t("input",{id:"metrics-disabled","onUpdate:modelValue":a[18]||(a[18]=s=>e.validate.metricsEnabled=s),value:"disabled",name:"metrics",type:"radio",class:"k-input mr-2"},null,512),[[E,e.validate.metricsEnabled]]),n(),Ue]),n(),t("label",Ae,[m(t("input",{id:"metrics-enabled","onUpdate:modelValue":a[19]||(a[19]=s=>e.validate.metricsEnabled=s),value:"enabled",name:"metrics",type:"radio",class:"k-input mr-2","data-testid":"mesh-metrics-enabled"},null,512),[[E,e.validate.metricsEnabled]]),n(),Be])]),_:1}),n(),e.validate.metricsEnabled==="enabled"?(u(),h(c,{key:0,class:"mt-4",title:"Backend name","for-attr":"metrics-name"},{default:l(()=>[m(t("input",{id:"metrics-name","onUpdate:modelValue":a[20]||(a[20]=s=>e.validate.meshMetricsName=s),type:"text",class:"k-input w-100",placeholder:"your-metrics-backend-name","data-testid":"mesh-metrics-backend-name"},null,512),[[f,e.validate.meshMetricsName]])]),_:1})):p("",!0),n(),e.validate.metricsEnabled==="enabled"?(u(),h(c,{key:1,class:"mt-4",title:"Type","for-attr":"metrics-type"},{default:l(()=>[m(t("select",{id:"metrics-type","onUpdate:modelValue":a[21]||(a[21]=s=>e.validate.meshMetricsType=s),class:"k-input w-100",name:"metrics-type"},Ve,512),[[U,e.validate.meshMetricsType]])]),_:1})):p("",!0),n(),e.validate.metricsEnabled==="enabled"?(u(),h(c,{key:2,class:"mt-4",title:"Dataplane port","for-attr":"metrics-dataplane-port"},{default:l(()=>[m(t("input",{id:"metrics-dataplane-port","onUpdate:modelValue":a[22]||(a[22]=s=>e.validate.meshMetricsDataplanePort=s),type:"number",class:"k-input w-100",step:"1",min:"0",max:"65535",placeholder:"1234"},null,512),[[f,e.validate.meshMetricsDataplanePort]])]),_:1})):p("",!0),n(),e.validate.metricsEnabled==="enabled"?(u(),h(c,{key:3,class:"mt-4",title:"Dataplane path","for-attr":"metrics-dataplane-path"},{default:l(()=>[m(t("input",{id:"metrics-dataplane-path","onUpdate:modelValue":a[23]||(a[23]=s=>e.validate.meshMetricsDataplanePath=s),type:"text",class:"k-input w-100"},null,512),[[f,e.validate.meshMetricsDataplanePath]])]),_:1})):p("",!0)]),_:1})]),complete:l(()=>[b.codeOutput?(u(),D("div",Ie,[e.hideScannerSiblings===!1?(u(),D("div",Re,[xe,n(),t("p",null,`
                Since the `+C(e.productName)+` GUI is read-only mode to follow Ops best practices,
                please execute the following command in your shell to create the entity.
                `+C(e.productName)+` will automatically detect when the new entity has been created.
              `,1),n(),g(w,{tabs:e.tabs,"initial-tab-override":i.environment,onOnTabChange:b.onTabChange},{kubernetes:l(()=>[g(y,{id:"code-block-kubernetes-command","data-testid":"kubernetes",language:"bash",code:b.codeOutput},null,8,["code"])]),universal:l(()=>[g(y,{id:"code-block-universal-command","data-testid":"universal",language:"bash",code:b.codeOutput},null,8,["code"])]),_:1},8,["tabs","initial-tab-override","onOnTabChange"])])):p("",!0),n(),g(r,{"loader-function":b.scanForEntity,"should-start":!0,"has-error":e.scanError,"can-complete":e.scanFound,onHideSiblings:b.hideSiblings},{"loading-title":l(()=>[Oe]),"loading-content":l(()=>[Pe]),"complete-title":l(()=>[Ke]),"complete-content":l(()=>[t("p",null,[n(`
                  Your mesh `),e.validate.meshName?(u(),D("strong",Fe,C(e.validate.meshName),1)):p("",!0),n(` was found!
                `)]),n(),t("p",null,[g(S,{appearance:"primary",to:{name:"mesh-detail-view",params:{mesh:e.validate.meshName}}},{default:l(()=>[n(`
                    Go to mesh `+C(e.validate.meshName),1)]),_:1},8,["to"])])]),"error-title":l(()=>[ze]),"error-content":l(()=>[je]),_:1},8,["loader-function","has-error","can-complete","onHideSiblings"])])):(u(),h(M,{key:1,appearance:"danger"},{alertMessage:l(()=>[Ye]),_:1}))]),mesh:l(()=>[We,n(),t("p",null,`
            In `+C(i.title)+`, a Mesh resource allows you to define an isolated environment
            for your data-planes and policies. It's isolated because the mTLS CA
            you choose can be different from the one configured for our Meshes.
            Ideally, you will have either a large Mesh with all the workloads, or
            one Mesh per application for better isolation.
          `,1),n(),t("p",null,[t("a",{href:`https://kuma.io/docs/${i.kumaDocsVersion}/policies/mesh/${e.utm}`,target:"_blank"},`
              Learn More
            `,8,Ge)])]),"did-you-know":l(()=>[Ze,n(),qe]),_:1},8,["steps","sidebar-content","footer-enabled","next-disabled","onGoToStep"])])])}const lt=V(Q,[["render",He],["__scopeId","data-v-7dc40dea"]]);export{lt as default};
