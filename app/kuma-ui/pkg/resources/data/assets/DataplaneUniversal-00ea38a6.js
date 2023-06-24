import{d as O,o as d,a as M,w as l,k as a,n as ue,h as r,b as u,g as e,v as de,q as f,c as S,s as F,t as m,D as h,G,e as w,j as B,F as W,J as j,E as b,f as q,p as pe,m as ce}from"./index-e8e69e62.js";import{l as ve,a as me,U as he,S as T,Z as _e}from"./kongponents.es-aa96ab2e.js";import{_ as fe}from"./EntityScanner.vue_vue_type_script_setup_true_lang-544e67ef.js";import{E as we}from"./EnvironmentSwitcher-c47bacfa.js";import{S as De,F as c}from"./StepSkeleton-6a73a6d6.js";import{h as R,j as ke,k as ge,f as ye,K as A,g as Se,_ as be}from"./RouteView.vue_vue_type_script_setup_true_lang-b28dd8e7.js";import{_ as Ne}from"./RouteTitle.vue_vue_type_script_setup_true_lang-8f9a216a.js";import{_ as U}from"./CodeBlock.vue_vue_type_style_index_0_lang-cc5f6631.js";import{Q as Pe}from"./QueryParameter-70743f73.js";import{t as Ie}from"./toYaml-4e00099e.js";const xe={type:"Dataplane",mesh:null,name:null,networking:{}},Te=O({__name:"HelperTooltip",props:{placement:{type:String,required:!1,default:void 0}},setup(_){const N=_;return(v,$)=>(d(),M(u(me),{class:"help-icon",trigger:"click",width:"100%","max-width":"300",placement:N.placement},{content:l(()=>[a("div",null,[ue(v.$slots,"default",{},void 0,!0)])]),default:l(()=>[r(u(ve),{icon:"info",color:"var(--grey-500)",size:"16","hide-title":""}),e()]),_:3},8,["placement"]))}});const D=R(Te,[["__scopeId","data-v-8cf92c6f"]]);function Ae(){return"https://localhost:5678/".replace("localhost",window.location.hostname)}const i=_=>(pe("data-v-b476f8ff"),_=_(),ce(),_),Ue={class:"wizard"},Me={class:"wizard__content"},$e=i(()=>a("h3",null,`
                Create Universal Dataplane
              `,-1)),Ce=i(()=>a("h3",null,`
                To get started, please select on what Mesh you would like to add the Dataplane:
              `,-1)),Ee=i(()=>a("p",null,`
                If you've got an existing Mesh that you would like to associate with your
                Dataplane, you can select it below, or create a new one using our Mesh Wizard.
              `,-1)),Ve=i(()=>a("small",null,"Would you like to see instructions for Kubernetes? Use sidebar to change wizard!",-1)),ze=i(()=>a("option",{disabled:"",value:""},`
                          Select an existing Mesh…
                        `,-1)),Fe=["value"],Ge=i(()=>a("label",{class:"k-input-label mr-4"},`
                        or
                      `,-1)),Be=i(()=>a("h3",null,`
                Setup Dataplane Mode
              `,-1)),We=i(()=>a("p",null,`
                You can create a data plane for a service or a data plane for a Gateway.
              `,-1)),je={for:"service-dataplane"},qe=i(()=>a("span",null,`
                      Service Dataplane
                    `,-1)),Oe={for:"gateway-dataplane"},Re=i(()=>a("span",null,`
                      Gateway Dataplane
                    `,-1)),He=["disabled"],Ke=i(()=>a("h3",null,`
                Networking
              `,-1)),Ye=i(()=>a("p",null,`
                It's time to now configure the networking settings so that the Dataplane
                can connect to the local service, and other data planes can consume
                your service.
              `,-1)),Ze=i(()=>a("p",null,[a("strong",null,"All fields below are required to proceed.")],-1)),Le=["value"],Qe=["value"],Je=["value","selected"],Xe={key:0},ea={key:0},aa=i(()=>a("h3",null,`
                    Auto-Inject DPP
                  `,-1)),ta=i(()=>a("h4",null,"Generate Dataplane Token",-1)),na=i(()=>a("h4",null,"Start Dataplane Process",-1)),la=i(()=>a("h3",null,"Searching…",-1)),oa=i(()=>a("p",null,"We are looking for your dataplane.",-1)),sa=i(()=>a("h3",null,"Done!",-1)),ra={key:0},ia=i(()=>a("p",null,`
                      Proceed to the next step where we will show you
                      your new Dataplane.
                    `,-1)),ua=i(()=>a("h3",null,"Dataplane not found",-1)),da=i(()=>a("p",null,"We were unable to find your dataplane.",-1)),pa=i(()=>a("p",null,`
                    Please return to the first step and make sure to select an
                    existing Mesh, or create a new one.
                  `,-1)),ca=i(()=>a("h3",null,"Dataplane",-1)),va=i(()=>a("h3",null,"Example",-1)),ma=i(()=>a("p",null,`
                Below is an example of a Dataplane resource output:
              `,-1)),ha=`type: Dataplane
mesh: default
name: dp-echo-1
networking:
  address: 10.0.0.1
  inbound:
  - port: 10000
    servicePort: 9000
    tags:
      kuma.io/service: echo`,_a=O({__name:"DataplaneUniversal",setup(_){const N=ke(),{t:v}=ge(),$=[{label:"General",slug:"general"},{label:"Topology",slug:"topology"},{label:"Networking",slug:"networking"},{label:"Install",slug:"complete"}],H=[{name:"dataplane"},{name:"example"},{name:"switch"}],K=Math.random().toString(36).substring(2,8),Y=de(),C=ye(),k=f(0),E=f(!1),P=f(!1),g=f(!1),V=f(!1),t=f({meshName:"",univDataplaneType:"dataplane-type-service",univDataplaneServiceName:"",univDataplaneId:"",univDataplaneCustomIdDisabled:!0,univDataplaneNetworkAddress:null,univDataplaneNetworkServicePort:null,univDataplaneNetworkServiceAddress:"127.0.0.1",univDataplaneNetworkDPPort:null,univDataplaneNetworkProtocol:"tcp"}),Z=f({protocols:["tcp","http","grpc"]}),L=S(()=>{const o=Object.assign({},xe),{meshName:s,univDataplaneType:n,univDataplaneServiceName:p,univDataplaneId:I,univDataplaneNetworkAddress:y,univDataplaneNetworkServicePort:x,univDataplaneNetworkServiceAddress:se,univDataplaneNetworkDPPort:re,univDataplaneNetworkProtocol:ie}=t.value;return s?(o.name=I,o.mesh=s,n==="dataplane-type-service"?(o.networking.gateway&&delete o.networking.gateway,o.networking={address:y,inbound:[{port:re,servicePort:x,serviceAddress:se,tags:{"kuma.io/service":p,"kuma.io/protocol":ie}}]}):n==="dataplane-type-gateway"&&(o.networking.inbound&&delete o.networking.inbound,o.networking={address:y,gateway:{tags:{"kuma.io/service":p}}}),o):""}),Q=S(()=>{const{univDataplaneId:o}=t.value;return`kumactl generate dataplane-token --name=${o} > kuma-token-${o}`}),J=S(()=>{const{univDataplaneId:o}=t.value;return`kuma-dp run \\
  --cp-address=${Ae()} \\
  --dataplane=${`"${Ie(L.value)}"`} \\
  --dataplane-token-file=kuma-token-${o}`}),X=S(()=>{const{meshName:o,univDataplaneServiceName:s,univDataplaneId:n,univDataplaneNetworkAddress:p,univDataplaneNetworkServicePort:I,univDataplaneNetworkDPPort:y,univDataplaneNetworkProtocol:x}=t.value;return o.length===0?!0:k.value===1?!(s&&n):k.value===2?!(p&&I&&y&&x):!1});F(()=>t.value.univDataplaneId,function(o){t.value.univDataplaneId=A(o)}),F(()=>t.value.univDataplaneServiceName,function(o){t.value.univDataplaneServiceName=A(o),t.value.univDataplaneServiceName===""?t.value.univDataplaneId="":t.value.univDataplaneId=A(`${o}-${K}`)});const z=Pe.get("step");k.value=z!==null?parseInt(z):0;function ee(o){k.value=o}function ae(){P.value=!0}async function te(){var n;const{meshName:o,univDataplaneId:s}=t.value;if(V.value=!1,g.value=!1,!(!o||!s))try{((n=(await N.getDataplaneFromMesh({mesh:o,name:s})).name)==null?void 0:n.length)>0?E.value=!0:g.value=!0}catch(p){g.value=!0,console.error(p)}finally{V.value=!0}}function ne(){C.dispatch("updateSelectedMesh",t.value.meshName),Y.push({name:"data-planes-list-view",params:{mesh:t.value.meshName}})}function le(o){const n=o.target.value.replace(/[a-zA-Z]*$/g,"").trim();t.value.univDataplaneNetworkDPPort=n===""?null:Number(n)}function oe(o){const n=o.target.value.replace(/[a-zA-Z]*$/g,"").trim();t.value.univDataplaneNetworkServicePort=n===""?null:Number(n)}return(o,s)=>(d(),M(be,null,{default:l(()=>[r(Ne,{title:u(v)("wizard-universal.routes.item.title")},null,8,["title"]),e(),r(Se,null,{default:l(()=>[a("div",Ue,[a("div",Me,[r(De,{steps:$,"sidebar-content":H,"footer-enabled":P.value===!1,"next-disabled":X.value,onGoToStep:ee},{general:l(()=>[$e,e(),a("p",null,`
                Welcome to the wizard to create a new Dataplane resource in `+m(u(v)("common.product.name"))+`.
                We will be providing you with a few steps that will get you started.
              `,1),e(),a("p",null,`
                As you know, the `+m(u(v)("common.product.name"))+` GUI is read-only.
              `,1),e(),Ce,e(),Ee,e(),Ve,e(),r(u(he),{class:"my-6","has-shadow":""},{body:l(()=>[r(c,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:l(()=>[a("div",null,[h(a("select",{id:"dp-mesh","onUpdate:modelValue":s[0]||(s[0]=n=>t.value.meshName=n),class:"k-input w-100","data-testid":"mesh-select"},[ze,e(),(d(!0),w(W,null,B(u(C).getters.getMeshList.items,n=>(d(),w("option",{key:n.name,value:n.name},m(n.name),9,Fe))),128))],512),[[G,t.value.meshName]])]),e(),a("div",null,[Ge,e(),r(u(T),{to:{name:"create-mesh"},appearance:"secondary"},{default:l(()=>[e(`
                        Create a new Mesh
                      `)]),_:1})])]),_:1})]),_:1})]),topology:l(()=>[Be,e(),We,e(),r(c,{"all-inline":"","equal-cols":"","hide-label-col":"","shift-right":""},{default:l(()=>[a("div",null,[a("label",je,[h(a("input",{id:"service-dataplane","onUpdate:modelValue":s[1]||(s[1]=n=>t.value.univDataplaneType=n),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[j,t.value.univDataplaneType]]),e(),qe]),e(),a("label",Oe,[h(a("input",{id:"gateway-dataplane","onUpdate:modelValue":s[2]||(s[2]=n=>t.value.univDataplaneType=n),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-gateway"},null,512),[[j,t.value.univDataplaneType]]),e(),Re])])]),_:1}),e(),r(c,{"all-inline":"",title:"Service name","for-attr":"service-name"},{default:l(()=>[h(a("input",{id:"service-name","onUpdate:modelValue":s[3]||(s[3]=n=>t.value.univDataplaneServiceName=n),"data-testid":"service-name",type:"text",class:"k-input w-100 mr-4"},null,512),[[b,t.value.univDataplaneServiceName]])]),_:1}),e(),r(c,{"all-inline":"",title:"Dataplane ID","for-attr":"dataplane-id"},{default:l(()=>[a("div",null,[h(a("input",{id:"dataplane-id","onUpdate:modelValue":s[4]||(s[4]=n=>t.value.univDataplaneId=n),type:"text",class:"k-input w-100",disabled:t.value.univDataplaneCustomIdDisabled,"data-testid":"dataplane-id"},null,8,He),[[b,t.value.univDataplaneId]])]),e(),a("div",null,[r(u(T),{appearance:"secondary","data-testid":"edit-button",onClick:s[5]||(s[5]=n=>t.value.univDataplaneCustomIdDisabled=!1)},{default:l(()=>[e(`
                    Edit
                  `)]),_:1}),e(),r(D,null,{default:l(()=>[e(`
                    This is a unique ID for the Dataplane instance.
                  `)]),_:1})])]),_:1})]),networking:l(()=>[Ke,e(),Ye,e(),Ze,e(),r(c,{"all-inline":"",title:"Data Plane IP Address","for-attr":"network-address"},{default:l(()=>[h(a("input",{id:"network-address","onUpdate:modelValue":s[6]||(s[6]=n=>t.value.univDataplaneNetworkAddress=n),type:"text",class:"k-input w-100","data-testid":"network-address"},null,512),[[b,t.value.univDataplaneNetworkAddress]]),e(),r(D,null,{default:l(()=>[e(`
                  The IP address that other services will use to consume this data plane.
                `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"","fill-first":"",title:"Data Plane Port","for-attr":"network-dataplane-port"},{default:l(()=>[a("input",{id:"network-dataplane-port",value:t.value.univDataplaneNetworkDPPort,type:"text",class:"k-input w-100","data-testid":"network-dataplane-port",onInput:le},null,40,Le),e(),r(D,null,{default:l(()=>[e(`
                  The data plane port (that other services will use to consume this service).
                `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"",title:"Service IP Address","for-attr":"network-service-address"},{default:l(()=>[h(a("input",{id:"network-service-address","onUpdate:modelValue":s[7]||(s[7]=n=>t.value.univDataplaneNetworkServiceAddress=n),type:"text",class:"k-input w-100"},null,512),[[b,t.value.univDataplaneNetworkServiceAddress]]),e(),r(D,null,{default:l(()=>[e(`
                  The address where your service is listening on the machine.
                `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"",title:"Service Port","for-attr":"network-service-port"},{default:l(()=>[a("input",{id:"network-service-port",value:t.value.univDataplaneNetworkServicePort,type:"text",class:"k-input w-100","data-testid":"service-port",onInput:oe},null,40,Qe),e(),r(D,null,{default:l(()=>[e(`
                  The port where your service is listening on the machine.
                `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"",title:"Protocol","for-attr":"network-dataplane-protocol"},{default:l(()=>[h(a("select",{id:"network-dataplane-protocol","onUpdate:modelValue":s[8]||(s[8]=n=>t.value.univDataplaneNetworkProtocol=n),class:"k-input w-100",name:"network-dataplane-protocol"},[(d(!0),w(W,null,B(Z.value.protocols,(n,p)=>(d(),w("option",{key:p,value:n,selected:t.value.univDataplaneNetworkProtocol===n},m(n),9,Je))),128))],512),[[G,t.value.univDataplaneNetworkProtocol]]),e(),r(D,null,{default:l(()=>[e(`
                  The protocol of the service.
                `)]),_:1})]),_:1})]),complete:l(()=>[t.value.meshName?(d(),w("div",Xe,[P.value===!1?(d(),w("div",ea,[aa,e(),a("p",null,`
                    It's time to first generate the credentials so that `+m(u(v)("common.product.name"))+` will allow
                    the Dataplane to successfully authenticate itself with the control plane,
                    and then finally install the Dataplane process (powered by Envoy).
                  `,1),e(),ta,e(),r(U,{id:"code-block-generate-token-command",language:"bash",code:Q.value},null,8,["code"]),e(),na,e(),r(U,{id:"code-block-stard-dp-command",language:"bash",code:J.value},null,8,["code"])])):q("",!0),e(),r(fe,{"loader-function":te,"has-error":g.value,"can-complete":E.value,onHideSiblings:ae},{"loading-title":l(()=>[la]),"loading-content":l(()=>[oa]),"complete-title":l(()=>[sa]),"complete-content":l(()=>[a("p",null,[e(`
                      Your Dataplane
                      `),t.value.univDataplaneId?(d(),w("strong",ra,m(t.value.univDataplaneId),1)):q("",!0),e(`
                      was found!
                    `)]),e(),ia,e(),a("p",null,[r(u(T),{appearance:"primary",onClick:ne},{default:l(()=>[e(`
                        View Your Dataplane
                      `)]),_:1})])]),"error-title":l(()=>[ua]),"error-content":l(()=>[da]),_:1},8,["has-error","can-complete"])])):(d(),M(u(_e),{key:1,appearance:"danger"},{alertMessage:l(()=>[pa]),_:1}))]),dataplane:l(()=>[ca,e(),a("p",null,`
                In `+m(u(v)("common.product.name"))+`, a Dataplane resource represents a data plane proxy running
                alongside one of your services. Data plane proxies can be added in any Mesh
                that you may have created, and in Kubernetes, they will be auto-injected
                by `+m(u(v)("common.product.name"))+`.
              `,1)]),example:l(()=>[va,e(),ma,e(),r(U,{id:"onboarding-dpp-universal-example",class:"sample-code-block mt-3",code:ha,language:"yaml"})]),switch:l(()=>[r(we)]),_:1},8,["footer-enabled","next-disabled"])])])]),_:1})]),_:1}))}});const Ia=R(_a,[["__scopeId","data-v-b476f8ff"]]);export{Ia as default};
