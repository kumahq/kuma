import{d as q,o as u,b as Z,w as l,i as a,z as ue,g as r,e as c,h as e,a as de,r as f,c as D,v as F,$ as M,j as v,t as m,P as pe,U as h,Z as G,F as B,q as O,a0 as R,V as N,f as W,p as ce,m as ve}from"./index-60be81af.js";import{s as me,c as he,Z as _e,x as U,V as fe}from"./kongponents.es-bdcd64fe.js";import{_ as we}from"./EntityScanner.vue_vue_type_script_setup_true_lang-e3ffa14e.js";import{E as De}from"./EnvironmentSwitcher-c2f41560.js";import{F as p,S as ke}from"./StepSkeleton-9f0ccc35.js";import{_ as j}from"./_plugin-vue_export-helper-c27b6911.js";import{_ as C}from"./CodeBlock.vue_vue_type_style_index_0_lang-7c7e3e4e.js";import{u as ge}from"./store-94b94e8a.js";import{u as ye}from"./index-c896859e.js";import{Q as Se}from"./QueryParameter-70743f73.js";import{t as Ne}from"./toYaml-4e00099e.js";const be={type:"Dataplane",mesh:null,name:null,networking:{}},Pe=q({__name:"HelperTooltip",props:{placement:{type:String,required:!1,default:void 0}},setup(_){const b=_;return(P,V)=>(u(),Z(c(he),{class:"help-icon",trigger:"click",width:"100%","max-width":"300",placement:b.placement},{content:l(()=>[a("div",null,[ue(P.$slots,"default",{},void 0,!0)])]),default:l(()=>[r(c(me),{icon:"info",color:"var(--grey-500)",size:"16","hide-title":""}),e()]),_:3},8,["placement"]))}});const w=j(Pe,[["__scopeId","data-v-8cf92c6f"]]);function Ie(){return"https://localhost:5678/".replace("localhost",window.location.hostname)}const i=_=>(ce("data-v-f5d99c85"),_=_(),ve(),_),xe={class:"wizard"},Te={class:"wizard__content"},Ae=i(()=>a("h3",null,`
            Create Universal Dataplane
          `,-1)),Me=i(()=>a("h3",null,`
            To get started, please select on what Mesh you would like to add the Dataplane:
          `,-1)),Ue=i(()=>a("p",null,`
            If you've got an existing Mesh that you would like to associate with your
            Dataplane, you can select it below, or create a new one using our Mesh Wizard.
          `,-1)),Ce=i(()=>a("small",null,"Would you like to see instructions for Kubernetes? Use sidebar to change wizard!",-1)),Ve=i(()=>a("option",{disabled:"",value:""},`
                      Select an existing Mesh…
                    `,-1)),$e=["value"],Ee=i(()=>a("label",{class:"k-input-label mr-4"},`
                    or
                  `,-1)),ze=i(()=>a("h3",null,`
            Setup Dataplane Mode
          `,-1)),Fe=i(()=>a("p",null,`
            You can create a data plane for a service or a data plane for a Gateway.
          `,-1)),Ge={for:"service-dataplane"},Be=i(()=>a("span",null,`
                  Service Dataplane
                `,-1)),Oe={for:"gateway-dataplane"},Re=i(()=>a("span",null,`
                  Gateway Dataplane
                `,-1)),We=["disabled"],qe=i(()=>a("h3",null,`
            Networking
          `,-1)),Ze=i(()=>a("p",null,`
            It's time to now configure the networking settings so that the Dataplane
            can connect to the local service, and other data planes can consume
            your service.
          `,-1)),je=i(()=>a("p",null,[a("strong",null,"All fields below are required to proceed.")],-1)),He=["value"],Ye=["value"],Ke=["value","selected"],Le={key:0},Qe={key:0},Xe=i(()=>a("h3",null,`
                Auto-Inject DPP
              `,-1)),Je=i(()=>a("h4",null,"Generate Dataplane Token",-1)),ea=i(()=>a("h4",null,"Start Dataplane Process",-1)),aa=i(()=>a("h3",null,"Searching…",-1)),ta=i(()=>a("p",null,"We are looking for your dataplane.",-1)),na=i(()=>a("h3",null,"Done!",-1)),la={key:0},oa=i(()=>a("p",null,`
                  Proceed to the next step where we will show you
                  your new Dataplane.
                `,-1)),sa=i(()=>a("h3",null,"Dataplane not found",-1)),ra=i(()=>a("p",null,"We were unable to find your dataplane.",-1)),ia=i(()=>a("p",null,`
                Please return to the first step and make sure to select an
                existing Mesh, or create a new one.
              `,-1)),ua=i(()=>a("h3",null,"Dataplane",-1)),da=i(()=>a("h3",null,"Example",-1)),pa=i(()=>a("p",null,`
            Below is an example of a Dataplane resource output:
          `,-1)),ca=`type: Dataplane
mesh: default
name: dp-echo-1
networking:
  address: 10.0.0.1
  inbound:
  - port: 10000
    servicePort: 9000
    tags:
      kuma.io/service: echo`,va=q({__name:"DataplaneUniversal",setup(_){const b=ye(),P=[{label:"General",slug:"general"},{label:"Topology",slug:"topology"},{label:"Networking",slug:"networking"},{label:"Install",slug:"complete"}],V=[{name:"dataplane"},{name:"example"},{name:"switch"}],H=Math.random().toString(36).substring(2,8),Y=de(),I=ge(),k=f(0),$=f(!1),x=f(!1),g=f(!1),E=f(!1),t=f({meshName:"",univDataplaneType:"dataplane-type-service",univDataplaneServiceName:"",univDataplaneId:"",univDataplaneCustomIdDisabled:!0,univDataplaneNetworkAddress:null,univDataplaneNetworkServicePort:null,univDataplaneNetworkServiceAddress:"127.0.0.1",univDataplaneNetworkDPPort:null,univDataplaneNetworkProtocol:"tcp"}),K=f({protocols:["tcp","http","grpc"]}),y=D(()=>I.getters["config/getTagline"]),L=D(()=>{const o=Object.assign({},be),{meshName:s,univDataplaneType:n,univDataplaneServiceName:d,univDataplaneId:T,univDataplaneNetworkAddress:S,univDataplaneNetworkServicePort:A,univDataplaneNetworkServiceAddress:se,univDataplaneNetworkDPPort:re,univDataplaneNetworkProtocol:ie}=t.value;return s?(o.name=T,o.mesh=s,n==="dataplane-type-service"?(o.networking.gateway&&delete o.networking.gateway,o.networking={address:S,inbound:[{port:re,servicePort:A,serviceAddress:se,tags:{"kuma.io/service":d,"kuma.io/protocol":ie}}]}):n==="dataplane-type-gateway"&&(o.networking.inbound&&delete o.networking.inbound,o.networking={address:S,gateway:{tags:{"kuma.io/service":d}}}),o):""}),Q=D(()=>{const{univDataplaneId:o}=t.value;return`kumactl generate dataplane-token --name=${o} > kuma-token-${o}`}),X=D(()=>{const{univDataplaneId:o}=t.value;return`kuma-dp run \\
  --cp-address=${Ie()} \\
  --dataplane=${`"${Ne(L.value)}"`} \\
  --dataplane-token-file=kuma-token-${o}`}),J=D(()=>{const{meshName:o,univDataplaneServiceName:s,univDataplaneId:n,univDataplaneNetworkAddress:d,univDataplaneNetworkServicePort:T,univDataplaneNetworkDPPort:S,univDataplaneNetworkProtocol:A}=t.value;return o.length===0?!0:k.value===1?!(s&&n):k.value===2?!(d&&T&&S&&A):!1});F(()=>t.value.univDataplaneId,function(o){t.value.univDataplaneId=M(o)}),F(()=>t.value.univDataplaneServiceName,function(o){t.value.univDataplaneServiceName=M(o),t.value.univDataplaneServiceName===""?t.value.univDataplaneId="":t.value.univDataplaneId=M(`${o}-${H}`)});const z=Se.get("step");k.value=z!==null?parseInt(z):0;function ee(o){k.value=o}function ae(){x.value=!0}async function te(){var n;const{meshName:o,univDataplaneId:s}=t.value;if(E.value=!1,g.value=!1,!(!o||!s))try{((n=(await b.getDataplaneFromMesh({mesh:o,name:s})).name)==null?void 0:n.length)>0?$.value=!0:g.value=!0}catch(d){g.value=!0,console.error(d)}finally{E.value=!0}}function ne(){I.dispatch("updateSelectedMesh",t.value.meshName),Y.push({name:"data-planes-list-view",params:{mesh:t.value.meshName}})}function le(o){const n=o.target.value.replace(/[a-zA-Z]*$/g,"").trim();t.value.univDataplaneNetworkDPPort=n===""?null:Number(n)}function oe(o){const n=o.target.value.replace(/[a-zA-Z]*$/g,"").trim();t.value.univDataplaneNetworkServicePort=n===""?null:Number(n)}return(o,s)=>(u(),v("div",xe,[a("div",Te,[r(ke,{steps:P,"sidebar-content":V,"footer-enabled":x.value===!1,"next-disabled":J.value,onGoToStep:ee},{general:l(()=>[Ae,e(),a("p",null,`
            Welcome to the wizard to create a new Dataplane resource in `+m(y.value)+`.
            We will be providing you with a few steps that will get you started.
          `,1),e(),a("p",null,`
            As you know, the `+m(c(pe))+` GUI is read-only.
          `,1),e(),Me,e(),Ue,e(),Ce,e(),r(c(_e),{class:"my-6","has-shadow":""},{body:l(()=>[r(p,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:l(()=>[a("div",null,[h(a("select",{id:"dp-mesh","onUpdate:modelValue":s[0]||(s[0]=n=>t.value.meshName=n),class:"k-input w-100","data-testid":"mesh-select"},[Ve,e(),(u(!0),v(B,null,O(c(I).getters.getMeshList.items,n=>(u(),v("option",{key:n.name,value:n.name},m(n.name),9,$e))),128))],512),[[G,t.value.meshName]])]),e(),a("div",null,[Ee,e(),r(c(U),{to:{name:"create-mesh"},appearance:"secondary"},{default:l(()=>[e(`
                    Create a new Mesh
                  `)]),_:1})])]),_:1})]),_:1})]),topology:l(()=>[ze,e(),Fe,e(),r(p,{"all-inline":"","equal-cols":"","hide-label-col":"","shift-right":""},{default:l(()=>[a("div",null,[a("label",Ge,[h(a("input",{id:"service-dataplane","onUpdate:modelValue":s[1]||(s[1]=n=>t.value.univDataplaneType=n),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[R,t.value.univDataplaneType]]),e(),Be]),e(),a("label",Oe,[h(a("input",{id:"gateway-dataplane","onUpdate:modelValue":s[2]||(s[2]=n=>t.value.univDataplaneType=n),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-gateway"},null,512),[[R,t.value.univDataplaneType]]),e(),Re])])]),_:1}),e(),r(p,{"all-inline":"",title:"Service name","for-attr":"service-name"},{default:l(()=>[h(a("input",{id:"service-name","onUpdate:modelValue":s[3]||(s[3]=n=>t.value.univDataplaneServiceName=n),"data-testid":"service-name",type:"text",class:"k-input w-100 mr-4"},null,512),[[N,t.value.univDataplaneServiceName]])]),_:1}),e(),r(p,{"all-inline":"",title:"Dataplane ID","for-attr":"dataplane-id"},{default:l(()=>[a("div",null,[h(a("input",{id:"dataplane-id","onUpdate:modelValue":s[4]||(s[4]=n=>t.value.univDataplaneId=n),type:"text",class:"k-input w-100",disabled:t.value.univDataplaneCustomIdDisabled,"data-testid":"dataplane-id"},null,8,We),[[N,t.value.univDataplaneId]])]),e(),a("div",null,[r(c(U),{appearance:"secondary","data-testid":"edit-button",onClick:s[5]||(s[5]=n=>t.value.univDataplaneCustomIdDisabled=!1)},{default:l(()=>[e(`
                Edit
              `)]),_:1}),e(),r(w,null,{default:l(()=>[e(`
                This is a unique ID for the Dataplane instance.
              `)]),_:1})])]),_:1})]),networking:l(()=>[qe,e(),Ze,e(),je,e(),r(p,{"all-inline":"",title:"Data Plane IP Address","for-attr":"network-address"},{default:l(()=>[h(a("input",{id:"network-address","onUpdate:modelValue":s[6]||(s[6]=n=>t.value.univDataplaneNetworkAddress=n),type:"text",class:"k-input w-100","data-testid":"network-address"},null,512),[[N,t.value.univDataplaneNetworkAddress]]),e(),r(w,null,{default:l(()=>[e(`
              The IP address that other services will use to consume this data plane.
            `)]),_:1})]),_:1}),e(),r(p,{"all-inline":"","fill-first":"",title:"Data Plane Port","for-attr":"network-dataplane-port"},{default:l(()=>[a("input",{id:"network-dataplane-port",value:t.value.univDataplaneNetworkDPPort,type:"text",class:"k-input w-100","data-testid":"network-dataplane-port",onInput:le},null,40,He),e(),r(w,null,{default:l(()=>[e(`
              The data plane port (that other services will use to consume this service).
            `)]),_:1})]),_:1}),e(),r(p,{"all-inline":"",title:"Service IP Address","for-attr":"network-service-address"},{default:l(()=>[h(a("input",{id:"network-service-address","onUpdate:modelValue":s[7]||(s[7]=n=>t.value.univDataplaneNetworkServiceAddress=n),type:"text",class:"k-input w-100"},null,512),[[N,t.value.univDataplaneNetworkServiceAddress]]),e(),r(w,null,{default:l(()=>[e(`
              The address where your service is listening on the machine.
            `)]),_:1})]),_:1}),e(),r(p,{"all-inline":"",title:"Service Port","for-attr":"network-service-port"},{default:l(()=>[a("input",{id:"network-service-port",value:t.value.univDataplaneNetworkServicePort,type:"text",class:"k-input w-100","data-testid":"service-port",onInput:oe},null,40,Ye),e(),r(w,null,{default:l(()=>[e(`
              The port where your service is listening on the machine.
            `)]),_:1})]),_:1}),e(),r(p,{"all-inline":"",title:"Protocol","for-attr":"network-dataplane-protocol"},{default:l(()=>[h(a("select",{id:"network-dataplane-protocol","onUpdate:modelValue":s[8]||(s[8]=n=>t.value.univDataplaneNetworkProtocol=n),class:"k-input w-100",name:"network-dataplane-protocol"},[(u(!0),v(B,null,O(K.value.protocols,(n,d)=>(u(),v("option",{key:d,value:n,selected:t.value.univDataplaneNetworkProtocol===n},m(n),9,Ke))),128))],512),[[G,t.value.univDataplaneNetworkProtocol]]),e(),r(w,null,{default:l(()=>[e(`
              The protocol of the service.
            `)]),_:1})]),_:1})]),complete:l(()=>[t.value.meshName?(u(),v("div",Le,[x.value===!1?(u(),v("div",Qe,[Xe,e(),a("p",null,`
                It's time to first generate the credentials so that `+m(y.value)+` will allow
                the Dataplane to successfully authenticate itself with the control plane,
                and then finally install the Dataplane process (powered by Envoy).
              `,1),e(),Je,e(),r(C,{id:"code-block-generate-token-command",language:"bash",code:Q.value},null,8,["code"]),e(),ea,e(),r(C,{id:"code-block-stard-dp-command",language:"bash",code:X.value},null,8,["code"])])):W("",!0),e(),r(we,{"loader-function":te,"has-error":g.value,"can-complete":$.value,onHideSiblings:ae},{"loading-title":l(()=>[aa]),"loading-content":l(()=>[ta]),"complete-title":l(()=>[na]),"complete-content":l(()=>[a("p",null,[e(`
                  Your Dataplane
                  `),t.value.univDataplaneId?(u(),v("strong",la,m(t.value.univDataplaneId),1)):W("",!0),e(`
                  was found!
                `)]),e(),oa,e(),a("p",null,[r(c(U),{appearance:"primary",onClick:ne},{default:l(()=>[e(`
                    View Your Dataplane
                  `)]),_:1})])]),"error-title":l(()=>[sa]),"error-content":l(()=>[ra]),_:1},8,["has-error","can-complete"])])):(u(),Z(c(fe),{key:1,appearance:"danger"},{alertMessage:l(()=>[ia]),_:1}))]),dataplane:l(()=>[ua,e(),a("p",null,`
            In `+m(y.value)+`, a Dataplane resource represents a data plane proxy running
            alongside one of your services. Data plane proxies can be added in any Mesh
            that you may have created, and in Kubernetes, they will be auto-injected
            by `+m(y.value)+`.
          `,1)]),example:l(()=>[da,e(),pa,e(),r(C,{id:"onboarding-dpp-universal-example",class:"sample-code-block mt-3",code:ca,language:"yaml"})]),switch:l(()=>[r(De)]),_:1},8,["footer-enabled","next-disabled"])])]))}});const ba=j(va,[["__scopeId","data-v-f5d99c85"]]);export{ba as default};
