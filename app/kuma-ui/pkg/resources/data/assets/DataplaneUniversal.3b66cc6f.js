import{d as B,o as v,c as E,w as o,u as O,j as t,K as z,b as e,cN as G,D as T,E as A,G as x,z as H,O as W,N as j,cl as q,co as L,cQ as S,k as R,h,i as w,a as r,t as D,l as c,cF as b,F as N,n as P,cR as I,v as f,A as C}from"./index.7a1b9f3b.js";import{j as Y}from"./index.58caa11d.js";import{k as Q}from"./kumaDpServerUrl.1e48e09f.js";import{_ as X}from"./CodeBlock.vue_vue_type_style_index_0_lang.9bb3d75c.js";import{F as Z,S as J,E as $}from"./EntityScanner.5f772ae6.js";import{E as ee}from"./EnvironmentSwitcher.a413c80c.js";import"./_commonjsHelpers.f037b798.js";const te=a=>(T("data-v-b945a6f3"),a=a(),A(),a),ae=te(()=>t("a",null,"?",-1)),ne=B({__name:"HelperTooltip",setup(a){return(n,d)=>(v(),E(O(G),{class:"help-icon",trigger:"hover"},{content:o(()=>[t("div",null,[z(n.$slots,"default",{},void 0,!0)])]),default:o(()=>[ae,e()]),_:3}))}});const le=x(ne,[["__scopeId","data-v-b945a6f3"]]),oe={type:"Dataplane",mesh:null,name:null,networking:{}};const ie=`type: Dataplane
mesh: default
name: dp-echo-1
networking:
  address: 10.0.0.1
  inbound:
  - port: 10000
    servicePort: 9000
    tags:
      kuma.io/service: echo`,se={name:"DataplaneWizardUniversal",EXAMPLE_CODE:ie,components:{CodeBlock:X,FormFragment:Z,StepSkeleton:J,EnvironmentSwitcher:ee,HelperTooltip:le,EntityScanner:$,KAlert:H,KButton:W,KCard:j},data(){return{productName:q,randString:Math.random().toString(36).substring(2,8),schema:oe,steps:[{label:"General",slug:"general"},{label:"Topology",slug:"topology"},{label:"Networking",slug:"networking"},{label:"Install",slug:"complete"}],tabs:[{hash:"#universal",title:"Universal"}],sidebarContent:[{name:"dataplane"},{name:"example"},{name:"switch"}],startScanner:!1,scanFound:!1,hideScannerSiblings:!1,scanError:!1,isComplete:!1,validate:{meshName:"",univDataplaneType:"dataplane-type-service",univDataplaneServiceName:"",univDataplaneId:"",univDataplaneCustomIdDisabled:!0,univDataplaneNetworkAddress:null,univDataplaneNetworkServicePort:null,univDataplaneNetworkServiceAddress:"127.0.0.1",univDataplaneNetworkDPPort:null,univDataplaneNetworkProtocol:"tcp"},formFields:{protocols:["tcp","http","grpc"]}}},computed:{...L({title:"config/getTagline",version:"config/getVersion",environment:"config/getEnvironment",meshes:"getMeshList"}),getDataplaneSchema(){const a=Object.assign({},this.schema),{meshName:n,univDataplaneType:d,univDataplaneServiceName:m,univDataplaneId:l,univDataplaneNetworkAddress:p,univDataplaneNetworkServicePort:k,univDataplaneNetworkServiceAddress:u,univDataplaneNetworkDPPort:g,univDataplaneNetworkProtocol:_}=this.validate;if(!!n)return a.name=l,a.mesh=n,d==="dataplane-type-service"?(a.networking.gateway&&delete a.networking.gateway,a.networking={address:p,inbound:[{port:g,servicePort:k,serviceAddress:u,tags:{"kuma.io/service":m,"kuma.io/protocol":_}}]}):d==="dataplane-type-gateway"&&(a.networking.inbound&&delete a.networking.inbound,a.networking={address:p,gateway:{tags:{"kuma.io/service":m}}}),a},generateDpTokenCodeOutput(){const{univDataplaneId:a}=this.validate;return`kumactl generate dataplane-token --name=${a} > kuma-token-${a}`},startDpCodeOutput(){const{univDataplaneId:a}=this.validate;return`kuma-dp run \\
      --cp-address=${Q()} \\
      --dataplane=${`"${Y(this.getDataplaneSchema)}"`} \\
      --dataplane-token-file=kuma-token-${a}`},nextDisabled(){const{meshName:a,univDataplaneServiceName:n,univDataplaneId:d,univDataplaneNetworkAddress:m,univDataplaneNetworkServicePort:l,univDataplaneNetworkDPPort:p,univDataplaneNetworkProtocol:k}=this.validate;return a.length?this.$route.query.step==="1"?!(n&&d):this.$route.query.step==="2"?!(m&&l&&p&&k):!1:!0}},watch:{"validate.univDataplaneId"(a){this.validate.univDataplaneId=S(a)},"validate.univDataplaneServiceName"(a){const n=S(a);this.validate.univDataplaneServiceName=n,this.validate.univDataplaneServiceName===""?this.validate.univDataplaneId="":this.validate.univDataplaneId=S(`${a}-${this.randString}`)},"validate.univDataplaneNetworkServicePort"(a){const n=a.replace(/[a-zA-Z]*$/g,"").trim();this.validate.univDataplaneNetworkServicePort=n},"validate.univDataplaneNetworkDPPort"(a){const n=a.replace(/[a-zA-Z]*$/g,"").trim();this.validate.univDataplaneNetworkDPPort=n}},methods:{hideSiblings(){this.hideScannerSiblings=!0},scanForEntity(){const{meshName:a,univDataplaneId:n}=this.validate;this.scanComplete=!1,this.scanError=!1,!(!a||!n)&&R.getDataplaneFromMesh({mesh:a,name:n}).then(d=>{var m;((m=d==null?void 0:d.name)==null?void 0:m.length)>0?(this.isRunning=!0,this.scanFound=!0):this.scanError=!0}).catch(d=>{this.scanError=!0,console.error(d)}).finally(()=>{this.scanComplete=!0})},compeleteDataPlaneSetup(){this.$store.dispatch("updateSelectedMesh",this.validate.meshName),this.$router.push({name:"data-plane-list-view",params:{mesh:this.validate.meshName}})}}},s=a=>(T("data-v-fcc70b3f"),a=a(),A(),a),re={class:"wizard"},de={class:"wizard__content"},pe=s(()=>t("h3",null,`
            Create Universal Dataplane
          `,-1)),ue=s(()=>t("h3",null,`
            To get started, please select on what Mesh you would like to add the Dataplane:
          `,-1)),ce=s(()=>t("p",null,`
            If you've got an existing Mesh that you would like to associate with your
            Dataplane, you can select it below, or create a new one using our Mesh Wizard.
          `,-1)),ve=s(()=>t("small",null,"Would you like to see instructions for Kubernetes? Use sidebar to change wizard!",-1)),he=s(()=>t("option",{disabled:"",value:""},`
                      Select an existing Mesh\u2026
                    `,-1)),me=["value"],_e=s(()=>t("label",{class:"k-input-label mr-4"},`
                    or
                  `,-1)),we=s(()=>t("h3",null,`
            Setup Dataplane Mode
          `,-1)),De=s(()=>t("p",null,`
            You can create a data plane for a service or a data plane for a Gateway.
          `,-1)),ke={for:"service-dataplane"},fe=s(()=>t("span",null,`
                  Service Dataplane
                `,-1)),ge={for:"gateway-dataplane"},ye=s(()=>t("span",null,`
                  Gateway Dataplane
                `,-1)),Se=["disabled"],be=s(()=>t("h3",null,`
            Networking
          `,-1)),Ne=s(()=>t("p",null,`
            It's time to now configure the networking settings so that the Dataplane
            can connect to the local service, and other data planes can consume
            your service.
          `,-1)),Pe=s(()=>t("p",null,[t("strong",null,"All fields below are required to proceed.")],-1)),Ie=["value","selected"],Ce={key:0},Ee={key:0},Te=s(()=>t("h3",null,`
                Auto-Inject DPP
              `,-1)),Ae=s(()=>t("h4",null,"Generate Dataplane Token",-1)),xe=s(()=>t("h4",null,"Start Dataplane Process",-1)),Me=s(()=>t("h3",null,"Searching\u2026",-1)),Ue=s(()=>t("p",null,"We are looking for your dataplane.",-1)),Fe=s(()=>t("h3",null,"Done!",-1)),Ve={key:0},Ke=s(()=>t("p",null,`
                  Proceed to the next step where we will show you
                  your new Dataplane.
                `,-1)),Be=s(()=>t("h3",null,"Dataplane not found",-1)),Oe=s(()=>t("p",null,"We were unable to find your dataplane.",-1)),ze=s(()=>t("p",null,`
                Please return to the first step and make sure to select an
                existing Mesh, or create a new one.
              `,-1)),Ge=s(()=>t("h3",null,"Dataplane",-1)),He=s(()=>t("h3",null,"Example",-1)),We=s(()=>t("p",null,`
            Below is an example of a Dataplane resource output:
          `,-1));function je(a,n,d,m,l,p){const k=h("KButton"),u=h("FormFragment"),g=h("KCard"),_=h("HelperTooltip"),y=h("CodeBlock"),M=h("EntityScanner"),U=h("KAlert"),F=h("EnvironmentSwitcher"),V=h("StepSkeleton");return v(),w("div",re,[t("div",de,[r(V,{steps:l.steps,"sidebar-content":l.sidebarContent,"footer-enabled":l.hideScannerSiblings===!1,"next-disabled":p.nextDisabled},{general:o(()=>[pe,e(),t("p",null,`
            Welcome to the wizard to create a new Dataplane resource in `+D(a.title)+`.
            We will be providing you with a few steps that will get you started.
          `,1),e(),t("p",null,`
            As you know, the `+D(l.productName)+` GUI is read-only.
          `,1),e(),ue,e(),ce,e(),ve,e(),r(g,{class:"my-6","has-shadow":""},{body:o(()=>[r(u,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:o(()=>[t("div",null,[c(t("select",{id:"dp-mesh","onUpdate:modelValue":n[0]||(n[0]=i=>l.validate.meshName=i),class:"k-input w-100","data-testid":"mesh-select"},[he,e(),(v(!0),w(N,null,P(a.meshes.items,i=>(v(),w("option",{key:i.name,value:i.name},D(i.name),9,me))),128))],512),[[b,l.validate.meshName]])]),e(),t("div",null,[_e,e(),r(k,{to:{name:"create-mesh"},appearance:"secondary"},{default:o(()=>[e(`
                    Create a new Mesh
                  `)]),_:1})])]),_:1})]),_:1})]),topology:o(()=>[we,e(),De,e(),r(u,{"all-inline":"","equal-cols":"","hide-label-col":"","shift-right":""},{default:o(()=>[t("div",null,[t("label",ke,[c(t("input",{id:"service-dataplane","onUpdate:modelValue":n[1]||(n[1]=i=>l.validate.univDataplaneType=i),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[I,l.validate.univDataplaneType]]),e(),fe]),e(),t("label",ge,[c(t("input",{id:"gateway-dataplane","onUpdate:modelValue":n[2]||(n[2]=i=>l.validate.univDataplaneType=i),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-gateway"},null,512),[[I,l.validate.univDataplaneType]]),e(),ye])])]),_:1}),e(),r(u,{"all-inline":"",title:"Service name","for-attr":"service-name"},{default:o(()=>[c(t("input",{id:"service-name","onUpdate:modelValue":n[3]||(n[3]=i=>l.validate.univDataplaneServiceName=i),"data-testid":"service-name",type:"text",class:"k-input w-100 mr-4"},null,512),[[f,l.validate.univDataplaneServiceName]])]),_:1}),e(),r(u,{"all-inline":"",title:"Dataplane ID","for-attr":"dataplane-id"},{default:o(()=>[t("div",null,[c(t("input",{id:"dataplane-id","onUpdate:modelValue":n[4]||(n[4]=i=>l.validate.univDataplaneId=i),type:"text",class:"k-input w-100",disabled:l.validate.univDataplaneCustomIdDisabled,"data-testid":"dataplane-id"},null,8,Se),[[f,l.validate.univDataplaneId]])]),e(),t("div",null,[r(k,{appearance:"secondary","data-testid":"edit-button",onClick:n[5]||(n[5]=i=>l.validate.univDataplaneCustomIdDisabled=!1)},{default:o(()=>[e(`
                Edit
              `)]),_:1}),e(),r(_,null,{default:o(()=>[e(`
                This is a unique ID for the Dataplane instance.
              `)]),_:1})])]),_:1})]),networking:o(()=>[be,e(),Ne,e(),Pe,e(),r(u,{"all-inline":"",title:"Data Plane IP Address","for-attr":"network-address"},{default:o(()=>[c(t("input",{id:"network-address","onUpdate:modelValue":n[6]||(n[6]=i=>l.validate.univDataplaneNetworkAddress=i),type:"text",class:"k-input w-100","data-testid":"network-address"},null,512),[[f,l.validate.univDataplaneNetworkAddress]]),e(),r(_,null,{default:o(()=>[e(`
              The IP address that other services will use to consume this data plane.
            `)]),_:1})]),_:1}),e(),r(u,{"all-inline":"","fill-first":"",title:"Data Plane Port","for-attr":"network-dataplane-port"},{default:o(()=>[c(t("input",{id:"network-dataplane-port","onUpdate:modelValue":n[7]||(n[7]=i=>l.validate.univDataplaneNetworkDPPort=i),type:"text",class:"k-input w-100","data-testid":"network-dataplane-port"},null,512),[[f,l.validate.univDataplaneNetworkDPPort]]),e(),r(_,null,{default:o(()=>[e(`
              The data plane port (that other services will use to consume this service).
            `)]),_:1})]),_:1}),e(),r(u,{"all-inline":"",title:"Service IP Address","for-attr":"network-service-address"},{default:o(()=>[c(t("input",{id:"network-service-address","onUpdate:modelValue":n[8]||(n[8]=i=>l.validate.univDataplaneNetworkServiceAddress=i),type:"text",class:"k-input w-100"},null,512),[[f,l.validate.univDataplaneNetworkServiceAddress]]),e(),r(_,null,{default:o(()=>[e(`
              The address where your service is listening on the machine.
            `)]),_:1})]),_:1}),e(),r(u,{"all-inline":"",title:"Service Port","for-attr":"network-service-port"},{default:o(()=>[c(t("input",{id:"network-service-port","onUpdate:modelValue":n[9]||(n[9]=i=>l.validate.univDataplaneNetworkServicePort=i),type:"text",class:"k-input w-100","data-testid":"service-port"},null,512),[[f,l.validate.univDataplaneNetworkServicePort]]),e(),r(_,null,{default:o(()=>[e(`
              The port where your service is listening on the machine.
            `)]),_:1})]),_:1}),e(),r(u,{"all-inline":"",title:"Protocol","for-attr":"network-dataplane-protocol"},{default:o(()=>[c(t("select",{id:"network-dataplane-protocol","onUpdate:modelValue":n[10]||(n[10]=i=>l.validate.univDataplaneNetworkProtocol=i),class:"k-input w-100",name:"network-dataplane-protocol"},[(v(!0),w(N,null,P(l.formFields.protocols,(i,K)=>(v(),w("option",{key:K,value:i,selected:l.validate.univDataplaneNetworkProtocol===i},D(i),9,Ie))),128))],512),[[b,l.validate.univDataplaneNetworkProtocol]]),e(),r(_,null,{default:o(()=>[e(`
              The protocol of the service.
            `)]),_:1})]),_:1})]),complete:o(()=>[l.validate.meshName?(v(),w("div",Ce,[l.hideScannerSiblings===!1?(v(),w("div",Ee,[Te,e(),t("p",null,`
                It's time to first generate the credentials so that `+D(a.title)+` will allow
                the Dataplane to successfully authenticate itself with the control plane,
                and then finally install the Dataplane process (powered by Envoy).
              `,1),e(),Ae,e(),r(y,{id:"code-block-generate-token-command",language:"bash",code:p.generateDpTokenCodeOutput},null,8,["code"]),e(),xe,e(),r(y,{id:"code-block-stard-dp-command",language:"bash",code:p.startDpCodeOutput},null,8,["code"])])):C("",!0),e(),r(M,{"loader-function":p.scanForEntity,"should-start":!0,"has-error":l.scanError,"can-complete":l.scanFound,onHideSiblings:p.hideSiblings},{"loading-title":o(()=>[Me]),"loading-content":o(()=>[Ue]),"complete-title":o(()=>[Fe]),"complete-content":o(()=>[t("p",null,[e(`
                  Your Dataplane
                  `),l.validate.univDataplaneId?(v(),w("strong",Ve,D(l.validate.univDataplaneId),1)):C("",!0),e(`
                  was found!
                `)]),e(),Ke,e(),t("p",null,[r(k,{appearance:"primary",onClick:p.compeleteDataPlaneSetup},{default:o(()=>[e(`
                    View Your Dataplane
                  `)]),_:1},8,["onClick"])])]),"error-title":o(()=>[Be]),"error-content":o(()=>[Oe]),_:1},8,["loader-function","has-error","can-complete","onHideSiblings"])])):(v(),E(U,{key:1,appearance:"danger"},{alertMessage:o(()=>[ze]),_:1}))]),dataplane:o(()=>[Ge,e(),t("p",null,`
            In `+D(a.title)+`, a Dataplane resource represents a data plane proxy running
            alongside one of your services. Data plane proxies can be added in any Mesh
            that you may have created, and in Kubernetes, they will be auto-injected
            by `+D(a.title)+`.
          `,1)]),example:o(()=>[He,e(),We,e(),r(y,{id:"onboarding-dpp-universal-example",class:"sample-code-block mt-3",code:a.$options.EXAMPLE_CODE,language:"yaml"},null,8,["code"])]),switch:o(()=>[r(F)]),_:1},8,["steps","sidebar-content","footer-enabled","next-disabled"])])])}const Je=x(se,[["render",je],["__scopeId","data-v-fcc70b3f"]]);export{Je as default};
