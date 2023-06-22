import{d as q,o as u,a as C,w as l,k as a,n as de,h as r,b as d,g as e,D as pe,q as _,c as D,s as G,_ as U,t as v,L as ce,R as m,X as R,e as f,j as B,F as O,$ as W,S as N,f as j,p as ve,m as me}from"./index-9d631905.js";import{l as he,a as _e,U as fe,S as M,Z as we}from"./kongponents.es-bba90403.js";import{_ as De}from"./EntityScanner.vue_vue_type_script_setup_true_lang-a2d31605.js";import{E as ke}from"./EnvironmentSwitcher-a2fcb2d8.js";import{S as ge,F as c}from"./StepSkeleton-d1f7d2c0.js";import{h as H,j as ye,k as Se,f as Ne,g as be,_ as Pe}from"./RouteView.vue_vue_type_script_setup_true_lang-76145142.js";import{_ as Ie}from"./RouteTitle.vue_vue_type_script_setup_true_lang-f639963c.js";import{_ as $}from"./CodeBlock.vue_vue_type_style_index_0_lang-9125ad7e.js";import{Q as xe}from"./QueryParameter-70743f73.js";import{t as Te}from"./toYaml-4e00099e.js";const Ae={type:"Dataplane",mesh:null,name:null,networking:{}},Ue=q({__name:"HelperTooltip",props:{placement:{type:String,required:!1,default:void 0}},setup(h){const b=h;return(P,E)=>(u(),C(d(_e),{class:"help-icon",trigger:"click",width:"100%","max-width":"300",placement:b.placement},{content:l(()=>[a("div",null,[de(P.$slots,"default",{},void 0,!0)])]),default:l(()=>[r(d(he),{icon:"info",color:"var(--grey-500)",size:"16","hide-title":""}),e()]),_:3},8,["placement"]))}});const w=H(Ue,[["__scopeId","data-v-8cf92c6f"]]);function Me(){return"https://localhost:5678/".replace("localhost",window.location.hostname)}const i=h=>(ve("data-v-9f023d47"),h=h(),me(),h),$e={class:"wizard"},Ce={class:"wizard__content"},Ee=i(()=>a("h3",null,`
                Create Universal Dataplane
              `,-1)),Ve=i(()=>a("h3",null,`
                To get started, please select on what Mesh you would like to add the Dataplane:
              `,-1)),ze=i(()=>a("p",null,`
                If you've got an existing Mesh that you would like to associate with your
                Dataplane, you can select it below, or create a new one using our Mesh Wizard.
              `,-1)),Fe=i(()=>a("small",null,"Would you like to see instructions for Kubernetes? Use sidebar to change wizard!",-1)),Ge=i(()=>a("option",{disabled:"",value:""},`
                          Select an existing Mesh…
                        `,-1)),Re=["value"],Be=i(()=>a("label",{class:"k-input-label mr-4"},`
                        or
                      `,-1)),Oe=i(()=>a("h3",null,`
                Setup Dataplane Mode
              `,-1)),We=i(()=>a("p",null,`
                You can create a data plane for a service or a data plane for a Gateway.
              `,-1)),je={for:"service-dataplane"},qe=i(()=>a("span",null,`
                      Service Dataplane
                    `,-1)),He={for:"gateway-dataplane"},Le=i(()=>a("span",null,`
                      Gateway Dataplane
                    `,-1)),Ye=["disabled"],Ze=i(()=>a("h3",null,`
                Networking
              `,-1)),Ke=i(()=>a("p",null,`
                It's time to now configure the networking settings so that the Dataplane
                can connect to the local service, and other data planes can consume
                your service.
              `,-1)),Qe=i(()=>a("p",null,[a("strong",null,"All fields below are required to proceed.")],-1)),Xe=["value"],Je=["value"],ea=["value","selected"],aa={key:0},ta={key:0},na=i(()=>a("h3",null,`
                    Auto-Inject DPP
                  `,-1)),la=i(()=>a("h4",null,"Generate Dataplane Token",-1)),oa=i(()=>a("h4",null,"Start Dataplane Process",-1)),sa=i(()=>a("h3",null,"Searching…",-1)),ra=i(()=>a("p",null,"We are looking for your dataplane.",-1)),ia=i(()=>a("h3",null,"Done!",-1)),ua={key:0},da=i(()=>a("p",null,`
                      Proceed to the next step where we will show you
                      your new Dataplane.
                    `,-1)),pa=i(()=>a("h3",null,"Dataplane not found",-1)),ca=i(()=>a("p",null,"We were unable to find your dataplane.",-1)),va=i(()=>a("p",null,`
                    Please return to the first step and make sure to select an
                    existing Mesh, or create a new one.
                  `,-1)),ma=i(()=>a("h3",null,"Dataplane",-1)),ha=i(()=>a("h3",null,"Example",-1)),_a=i(()=>a("p",null,`
                Below is an example of a Dataplane resource output:
              `,-1)),fa=`type: Dataplane
mesh: default
name: dp-echo-1
networking:
  address: 10.0.0.1
  inbound:
  - port: 10000
    servicePort: 9000
    tags:
      kuma.io/service: echo`,wa=q({__name:"DataplaneUniversal",setup(h){const b=ye(),{t:P}=Se(),E=[{label:"General",slug:"general"},{label:"Topology",slug:"topology"},{label:"Networking",slug:"networking"},{label:"Install",slug:"complete"}],L=[{name:"dataplane"},{name:"example"},{name:"switch"}],Y=Math.random().toString(36).substring(2,8),Z=pe(),I=Ne(),k=_(0),V=_(!1),x=_(!1),g=_(!1),z=_(!1),t=_({meshName:"",univDataplaneType:"dataplane-type-service",univDataplaneServiceName:"",univDataplaneId:"",univDataplaneCustomIdDisabled:!0,univDataplaneNetworkAddress:null,univDataplaneNetworkServicePort:null,univDataplaneNetworkServiceAddress:"127.0.0.1",univDataplaneNetworkDPPort:null,univDataplaneNetworkProtocol:"tcp"}),K=_({protocols:["tcp","http","grpc"]}),y=D(()=>I.getters["config/getTagline"]),Q=D(()=>{const o=Object.assign({},Ae),{meshName:s,univDataplaneType:n,univDataplaneServiceName:p,univDataplaneId:T,univDataplaneNetworkAddress:S,univDataplaneNetworkServicePort:A,univDataplaneNetworkServiceAddress:re,univDataplaneNetworkDPPort:ie,univDataplaneNetworkProtocol:ue}=t.value;return s?(o.name=T,o.mesh=s,n==="dataplane-type-service"?(o.networking.gateway&&delete o.networking.gateway,o.networking={address:S,inbound:[{port:ie,servicePort:A,serviceAddress:re,tags:{"kuma.io/service":p,"kuma.io/protocol":ue}}]}):n==="dataplane-type-gateway"&&(o.networking.inbound&&delete o.networking.inbound,o.networking={address:S,gateway:{tags:{"kuma.io/service":p}}}),o):""}),X=D(()=>{const{univDataplaneId:o}=t.value;return`kumactl generate dataplane-token --name=${o} > kuma-token-${o}`}),J=D(()=>{const{univDataplaneId:o}=t.value;return`kuma-dp run \\
  --cp-address=${Me()} \\
  --dataplane=${`"${Te(Q.value)}"`} \\
  --dataplane-token-file=kuma-token-${o}`}),ee=D(()=>{const{meshName:o,univDataplaneServiceName:s,univDataplaneId:n,univDataplaneNetworkAddress:p,univDataplaneNetworkServicePort:T,univDataplaneNetworkDPPort:S,univDataplaneNetworkProtocol:A}=t.value;return o.length===0?!0:k.value===1?!(s&&n):k.value===2?!(p&&T&&S&&A):!1});G(()=>t.value.univDataplaneId,function(o){t.value.univDataplaneId=U(o)}),G(()=>t.value.univDataplaneServiceName,function(o){t.value.univDataplaneServiceName=U(o),t.value.univDataplaneServiceName===""?t.value.univDataplaneId="":t.value.univDataplaneId=U(`${o}-${Y}`)});const F=xe.get("step");k.value=F!==null?parseInt(F):0;function ae(o){k.value=o}function te(){x.value=!0}async function ne(){var n;const{meshName:o,univDataplaneId:s}=t.value;if(z.value=!1,g.value=!1,!(!o||!s))try{((n=(await b.getDataplaneFromMesh({mesh:o,name:s})).name)==null?void 0:n.length)>0?V.value=!0:g.value=!0}catch(p){g.value=!0,console.error(p)}finally{z.value=!0}}function le(){I.dispatch("updateSelectedMesh",t.value.meshName),Z.push({name:"data-planes-list-view",params:{mesh:t.value.meshName}})}function oe(o){const n=o.target.value.replace(/[a-zA-Z]*$/g,"").trim();t.value.univDataplaneNetworkDPPort=n===""?null:Number(n)}function se(o){const n=o.target.value.replace(/[a-zA-Z]*$/g,"").trim();t.value.univDataplaneNetworkServicePort=n===""?null:Number(n)}return(o,s)=>(u(),C(Pe,null,{default:l(()=>[r(Ie,{title:d(P)("wizard-universal.routes.item.title")},null,8,["title"]),e(),r(be,null,{default:l(()=>[a("div",$e,[a("div",Ce,[r(ge,{steps:E,"sidebar-content":L,"footer-enabled":x.value===!1,"next-disabled":ee.value,onGoToStep:ae},{general:l(()=>[Ee,e(),a("p",null,`
                Welcome to the wizard to create a new Dataplane resource in `+v(y.value)+`.
                We will be providing you with a few steps that will get you started.
              `,1),e(),a("p",null,`
                As you know, the `+v(d(ce))+` GUI is read-only.
              `,1),e(),Ve,e(),ze,e(),Fe,e(),r(d(fe),{class:"my-6","has-shadow":""},{body:l(()=>[r(c,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:l(()=>[a("div",null,[m(a("select",{id:"dp-mesh","onUpdate:modelValue":s[0]||(s[0]=n=>t.value.meshName=n),class:"k-input w-100","data-testid":"mesh-select"},[Ge,e(),(u(!0),f(O,null,B(d(I).getters.getMeshList.items,n=>(u(),f("option",{key:n.name,value:n.name},v(n.name),9,Re))),128))],512),[[R,t.value.meshName]])]),e(),a("div",null,[Be,e(),r(d(M),{to:{name:"create-mesh"},appearance:"secondary"},{default:l(()=>[e(`
                        Create a new Mesh
                      `)]),_:1})])]),_:1})]),_:1})]),topology:l(()=>[Oe,e(),We,e(),r(c,{"all-inline":"","equal-cols":"","hide-label-col":"","shift-right":""},{default:l(()=>[a("div",null,[a("label",je,[m(a("input",{id:"service-dataplane","onUpdate:modelValue":s[1]||(s[1]=n=>t.value.univDataplaneType=n),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[W,t.value.univDataplaneType]]),e(),qe]),e(),a("label",He,[m(a("input",{id:"gateway-dataplane","onUpdate:modelValue":s[2]||(s[2]=n=>t.value.univDataplaneType=n),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-gateway"},null,512),[[W,t.value.univDataplaneType]]),e(),Le])])]),_:1}),e(),r(c,{"all-inline":"",title:"Service name","for-attr":"service-name"},{default:l(()=>[m(a("input",{id:"service-name","onUpdate:modelValue":s[3]||(s[3]=n=>t.value.univDataplaneServiceName=n),"data-testid":"service-name",type:"text",class:"k-input w-100 mr-4"},null,512),[[N,t.value.univDataplaneServiceName]])]),_:1}),e(),r(c,{"all-inline":"",title:"Dataplane ID","for-attr":"dataplane-id"},{default:l(()=>[a("div",null,[m(a("input",{id:"dataplane-id","onUpdate:modelValue":s[4]||(s[4]=n=>t.value.univDataplaneId=n),type:"text",class:"k-input w-100",disabled:t.value.univDataplaneCustomIdDisabled,"data-testid":"dataplane-id"},null,8,Ye),[[N,t.value.univDataplaneId]])]),e(),a("div",null,[r(d(M),{appearance:"secondary","data-testid":"edit-button",onClick:s[5]||(s[5]=n=>t.value.univDataplaneCustomIdDisabled=!1)},{default:l(()=>[e(`
                    Edit
                  `)]),_:1}),e(),r(w,null,{default:l(()=>[e(`
                    This is a unique ID for the Dataplane instance.
                  `)]),_:1})])]),_:1})]),networking:l(()=>[Ze,e(),Ke,e(),Qe,e(),r(c,{"all-inline":"",title:"Data Plane IP Address","for-attr":"network-address"},{default:l(()=>[m(a("input",{id:"network-address","onUpdate:modelValue":s[6]||(s[6]=n=>t.value.univDataplaneNetworkAddress=n),type:"text",class:"k-input w-100","data-testid":"network-address"},null,512),[[N,t.value.univDataplaneNetworkAddress]]),e(),r(w,null,{default:l(()=>[e(`
                  The IP address that other services will use to consume this data plane.
                `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"","fill-first":"",title:"Data Plane Port","for-attr":"network-dataplane-port"},{default:l(()=>[a("input",{id:"network-dataplane-port",value:t.value.univDataplaneNetworkDPPort,type:"text",class:"k-input w-100","data-testid":"network-dataplane-port",onInput:oe},null,40,Xe),e(),r(w,null,{default:l(()=>[e(`
                  The data plane port (that other services will use to consume this service).
                `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"",title:"Service IP Address","for-attr":"network-service-address"},{default:l(()=>[m(a("input",{id:"network-service-address","onUpdate:modelValue":s[7]||(s[7]=n=>t.value.univDataplaneNetworkServiceAddress=n),type:"text",class:"k-input w-100"},null,512),[[N,t.value.univDataplaneNetworkServiceAddress]]),e(),r(w,null,{default:l(()=>[e(`
                  The address where your service is listening on the machine.
                `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"",title:"Service Port","for-attr":"network-service-port"},{default:l(()=>[a("input",{id:"network-service-port",value:t.value.univDataplaneNetworkServicePort,type:"text",class:"k-input w-100","data-testid":"service-port",onInput:se},null,40,Je),e(),r(w,null,{default:l(()=>[e(`
                  The port where your service is listening on the machine.
                `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"",title:"Protocol","for-attr":"network-dataplane-protocol"},{default:l(()=>[m(a("select",{id:"network-dataplane-protocol","onUpdate:modelValue":s[8]||(s[8]=n=>t.value.univDataplaneNetworkProtocol=n),class:"k-input w-100",name:"network-dataplane-protocol"},[(u(!0),f(O,null,B(K.value.protocols,(n,p)=>(u(),f("option",{key:p,value:n,selected:t.value.univDataplaneNetworkProtocol===n},v(n),9,ea))),128))],512),[[R,t.value.univDataplaneNetworkProtocol]]),e(),r(w,null,{default:l(()=>[e(`
                  The protocol of the service.
                `)]),_:1})]),_:1})]),complete:l(()=>[t.value.meshName?(u(),f("div",aa,[x.value===!1?(u(),f("div",ta,[na,e(),a("p",null,`
                    It's time to first generate the credentials so that `+v(y.value)+` will allow
                    the Dataplane to successfully authenticate itself with the control plane,
                    and then finally install the Dataplane process (powered by Envoy).
                  `,1),e(),la,e(),r($,{id:"code-block-generate-token-command",language:"bash",code:X.value},null,8,["code"]),e(),oa,e(),r($,{id:"code-block-stard-dp-command",language:"bash",code:J.value},null,8,["code"])])):j("",!0),e(),r(De,{"loader-function":ne,"has-error":g.value,"can-complete":V.value,onHideSiblings:te},{"loading-title":l(()=>[sa]),"loading-content":l(()=>[ra]),"complete-title":l(()=>[ia]),"complete-content":l(()=>[a("p",null,[e(`
                      Your Dataplane
                      `),t.value.univDataplaneId?(u(),f("strong",ua,v(t.value.univDataplaneId),1)):j("",!0),e(`
                      was found!
                    `)]),e(),da,e(),a("p",null,[r(d(M),{appearance:"primary",onClick:le},{default:l(()=>[e(`
                        View Your Dataplane
                      `)]),_:1})])]),"error-title":l(()=>[pa]),"error-content":l(()=>[ca]),_:1},8,["has-error","can-complete"])])):(u(),C(d(we),{key:1,appearance:"danger"},{alertMessage:l(()=>[va]),_:1}))]),dataplane:l(()=>[ma,e(),a("p",null,`
                In `+v(y.value)+`, a Dataplane resource represents a data plane proxy running
                alongside one of your services. Data plane proxies can be added in any Mesh
                that you may have created, and in Kubernetes, they will be auto-injected
                by `+v(y.value)+`.
              `,1)]),example:l(()=>[ha,e(),_a,e(),r($,{id:"onboarding-dpp-universal-example",class:"sample-code-block mt-3",code:fa,language:"yaml"})]),switch:l(()=>[r(ke)]),_:1},8,["footer-enabled","next-disabled"])])])]),_:1})]),_:1}))}});const Ta=H(wa,[["__scopeId","data-v-9f023d47"]]);export{Ta as default};
