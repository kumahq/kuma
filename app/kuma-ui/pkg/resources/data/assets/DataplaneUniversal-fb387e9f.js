import{d as q,o as d,e as j,w as l,i as a,K as de,g as r,b as u,h as e,a as pe,r as f,c as D,v as z,j as v,ak as M,t as m,P as ce,a0 as h,a3 as G,F as B,q as O,al as R,a1 as N,f as W,p as ve,m as me}from"./index-bb404c7e.js";import{d as he,f as _e,x as fe,M as C,I as we}from"./kongponents.es-3141fcfd.js";import{S as De,F as c,E as ke}from"./StepSkeleton-fe7343f7.js";import{E as ge}from"./EnvironmentSwitcher-47da7f3a.js";import{_ as H}from"./_plugin-vue_export-helper-c27b6911.js";import{_ as U}from"./CodeBlock.vue_vue_type_style_index_0_lang-7ab44394.js";import{u as ye}from"./store-674ffdd1.js";import{u as Se}from"./index-cb99abc8.js";import{Q as Ne}from"./QueryParameter-70743f73.js";import{t as be}from"./toYaml-4e00099e.js";const Pe={type:"Dataplane",mesh:null,name:null,networking:{}},Ie=q({__name:"HelperTooltip",props:{placement:{type:String,required:!1,default:void 0}},setup(_){const b=_;return(P,E)=>(d(),j(u(_e),{class:"help-icon",trigger:"click",width:"100%","max-width":"300",placement:b.placement},{content:l(()=>[a("div",null,[de(P.$slots,"default",{},void 0,!0)])]),default:l(()=>[r(u(he),{icon:"info",color:"var(--grey-500)",size:"16","hide-title":""}),e()]),_:3},8,["placement"]))}});const w=H(Ie,[["__scopeId","data-v-8cf92c6f"]]);function xe(){return"https://localhost:5678/".replace("localhost",window.location.hostname)}const i=_=>(ve("data-v-fe363841"),_=_(),me(),_),Te={class:"wizard"},Ae={class:"wizard__content"},Me=i(()=>a("h3",null,`
            Create Universal Dataplane
          `,-1)),Ce=i(()=>a("h3",null,`
            To get started, please select on what Mesh you would like to add the Dataplane:
          `,-1)),Ue=i(()=>a("p",null,`
            If you've got an existing Mesh that you would like to associate with your
            Dataplane, you can select it below, or create a new one using our Mesh Wizard.
          `,-1)),Ee=i(()=>a("small",null,"Would you like to see instructions for Kubernetes? Use sidebar to change wizard!",-1)),$e=i(()=>a("option",{disabled:"",value:""},`
                      Select an existing Mesh…
                    `,-1)),Ve=["value"],Fe=i(()=>a("label",{class:"k-input-label mr-4"},`
                    or
                  `,-1)),ze=i(()=>a("h3",null,`
            Setup Dataplane Mode
          `,-1)),Ge=i(()=>a("p",null,`
            You can create a data plane for a service or a data plane for a Gateway.
          `,-1)),Be={for:"service-dataplane"},Oe=i(()=>a("span",null,`
                  Service Dataplane
                `,-1)),Re={for:"gateway-dataplane"},We=i(()=>a("span",null,`
                  Gateway Dataplane
                `,-1)),qe=["disabled"],je=i(()=>a("h3",null,`
            Networking
          `,-1)),He=i(()=>a("p",null,`
            It's time to now configure the networking settings so that the Dataplane
            can connect to the local service, and other data planes can consume
            your service.
          `,-1)),Ke=i(()=>a("p",null,[a("strong",null,"All fields below are required to proceed.")],-1)),Ye=["value"],Le=["value"],Qe=["value","selected"],Ze={key:0},Xe={key:0},Je=i(()=>a("h3",null,`
                Auto-Inject DPP
              `,-1)),ea=i(()=>a("h4",null,"Generate Dataplane Token",-1)),aa=i(()=>a("h4",null,"Start Dataplane Process",-1)),ta=i(()=>a("h3",null,"Searching…",-1)),na=i(()=>a("p",null,"We are looking for your dataplane.",-1)),la=i(()=>a("h3",null,"Done!",-1)),oa={key:0},sa=i(()=>a("p",null,`
                  Proceed to the next step where we will show you
                  your new Dataplane.
                `,-1)),ra=i(()=>a("h3",null,"Dataplane not found",-1)),ia=i(()=>a("p",null,"We were unable to find your dataplane.",-1)),ua=i(()=>a("p",null,`
                Please return to the first step and make sure to select an
                existing Mesh, or create a new one.
              `,-1)),da=i(()=>a("h3",null,"Dataplane",-1)),pa=i(()=>a("h3",null,"Example",-1)),ca=i(()=>a("p",null,`
            Below is an example of a Dataplane resource output:
          `,-1)),va=q({__name:"DataplaneUniversal",setup(_){const b=Se(),P=`type: Dataplane
mesh: default
name: dp-echo-1
networking:
  address: 10.0.0.1
  inbound:
  - port: 10000
    servicePort: 9000
    tags:
      kuma.io/service: echo`,E=[{label:"General",slug:"general"},{label:"Topology",slug:"topology"},{label:"Networking",slug:"networking"},{label:"Install",slug:"complete"}],K=[{name:"dataplane"},{name:"example"},{name:"switch"}],Y=Math.random().toString(36).substring(2,8),L=pe(),I=ye(),k=f(0),$=f(!1),x=f(!1),g=f(!1),V=f(!1),t=f({meshName:"",univDataplaneType:"dataplane-type-service",univDataplaneServiceName:"",univDataplaneId:"",univDataplaneCustomIdDisabled:!0,univDataplaneNetworkAddress:null,univDataplaneNetworkServicePort:null,univDataplaneNetworkServiceAddress:"127.0.0.1",univDataplaneNetworkDPPort:null,univDataplaneNetworkProtocol:"tcp"}),Q=f({protocols:["tcp","http","grpc"]}),y=D(()=>I.getters["config/getTagline"]),Z=D(()=>{const o=Object.assign({},Pe),{meshName:s,univDataplaneType:n,univDataplaneServiceName:p,univDataplaneId:T,univDataplaneNetworkAddress:S,univDataplaneNetworkServicePort:A,univDataplaneNetworkServiceAddress:re,univDataplaneNetworkDPPort:ie,univDataplaneNetworkProtocol:ue}=t.value;return s?(o.name=T,o.mesh=s,n==="dataplane-type-service"?(o.networking.gateway&&delete o.networking.gateway,o.networking={address:S,inbound:[{port:ie,servicePort:A,serviceAddress:re,tags:{"kuma.io/service":p,"kuma.io/protocol":ue}}]}):n==="dataplane-type-gateway"&&(o.networking.inbound&&delete o.networking.inbound,o.networking={address:S,gateway:{tags:{"kuma.io/service":p}}}),o):""}),X=D(()=>{const{univDataplaneId:o}=t.value;return`kumactl generate dataplane-token --name=${o} > kuma-token-${o}`}),J=D(()=>{const{univDataplaneId:o}=t.value;return`kuma-dp run \\
  --cp-address=${xe()} \\
  --dataplane=${`"${be(Z.value)}"`} \\
  --dataplane-token-file=kuma-token-${o}`}),ee=D(()=>{const{meshName:o,univDataplaneServiceName:s,univDataplaneId:n,univDataplaneNetworkAddress:p,univDataplaneNetworkServicePort:T,univDataplaneNetworkDPPort:S,univDataplaneNetworkProtocol:A}=t.value;return o.length===0?!0:k.value===1?!(s&&n):k.value===2?!(p&&T&&S&&A):!1});z(()=>t.value.univDataplaneId,function(o){t.value.univDataplaneId=M(o)}),z(()=>t.value.univDataplaneServiceName,function(o){t.value.univDataplaneServiceName=M(o),t.value.univDataplaneServiceName===""?t.value.univDataplaneId="":t.value.univDataplaneId=M(`${o}-${Y}`)});const F=Ne.get("step");k.value=F!==null?parseInt(F):0;function ae(o){k.value=o}function te(){x.value=!0}async function ne(){var n;const{meshName:o,univDataplaneId:s}=t.value;if(V.value=!1,g.value=!1,!(!o||!s))try{((n=(await b.getDataplaneFromMesh({mesh:o,name:s})).name)==null?void 0:n.length)>0?$.value=!0:g.value=!0}catch(p){g.value=!0,console.error(p)}finally{V.value=!0}}function le(){I.dispatch("updateSelectedMesh",t.value.meshName),L.push({name:"data-plane-list-view",params:{mesh:t.value.meshName}})}function oe(o){const n=o.target.value.replace(/[a-zA-Z]*$/g,"").trim();t.value.univDataplaneNetworkDPPort=n===""?null:Number(n)}function se(o){const n=o.target.value.replace(/[a-zA-Z]*$/g,"").trim();t.value.univDataplaneNetworkServicePort=n===""?null:Number(n)}return(o,s)=>(d(),v("div",Te,[a("div",Ae,[r(De,{steps:E,"sidebar-content":K,"footer-enabled":x.value===!1,"next-disabled":u(ee),onGoToStep:ae},{general:l(()=>[Me,e(),a("p",null,`
            Welcome to the wizard to create a new Dataplane resource in `+m(u(y))+`.
            We will be providing you with a few steps that will get you started.
          `,1),e(),a("p",null,`
            As you know, the `+m(u(ce))+` GUI is read-only.
          `,1),e(),Ce,e(),Ue,e(),Ee,e(),r(u(fe),{class:"my-6","has-shadow":""},{body:l(()=>[r(c,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:l(()=>[a("div",null,[h(a("select",{id:"dp-mesh","onUpdate:modelValue":s[0]||(s[0]=n=>t.value.meshName=n),class:"k-input w-100","data-testid":"mesh-select"},[$e,e(),(d(!0),v(B,null,O(u(I).getters.getMeshList.items,n=>(d(),v("option",{key:n.name,value:n.name},m(n.name),9,Ve))),128))],512),[[G,t.value.meshName]])]),e(),a("div",null,[Fe,e(),r(u(C),{to:{name:"create-mesh"},appearance:"secondary"},{default:l(()=>[e(`
                    Create a new Mesh
                  `)]),_:1})])]),_:1})]),_:1})]),topology:l(()=>[ze,e(),Ge,e(),r(c,{"all-inline":"","equal-cols":"","hide-label-col":"","shift-right":""},{default:l(()=>[a("div",null,[a("label",Be,[h(a("input",{id:"service-dataplane","onUpdate:modelValue":s[1]||(s[1]=n=>t.value.univDataplaneType=n),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[R,t.value.univDataplaneType]]),e(),Oe]),e(),a("label",Re,[h(a("input",{id:"gateway-dataplane","onUpdate:modelValue":s[2]||(s[2]=n=>t.value.univDataplaneType=n),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-gateway"},null,512),[[R,t.value.univDataplaneType]]),e(),We])])]),_:1}),e(),r(c,{"all-inline":"",title:"Service name","for-attr":"service-name"},{default:l(()=>[h(a("input",{id:"service-name","onUpdate:modelValue":s[3]||(s[3]=n=>t.value.univDataplaneServiceName=n),"data-testid":"service-name",type:"text",class:"k-input w-100 mr-4"},null,512),[[N,t.value.univDataplaneServiceName]])]),_:1}),e(),r(c,{"all-inline":"",title:"Dataplane ID","for-attr":"dataplane-id"},{default:l(()=>[a("div",null,[h(a("input",{id:"dataplane-id","onUpdate:modelValue":s[4]||(s[4]=n=>t.value.univDataplaneId=n),type:"text",class:"k-input w-100",disabled:t.value.univDataplaneCustomIdDisabled,"data-testid":"dataplane-id"},null,8,qe),[[N,t.value.univDataplaneId]])]),e(),a("div",null,[r(u(C),{appearance:"secondary","data-testid":"edit-button",onClick:s[5]||(s[5]=n=>t.value.univDataplaneCustomIdDisabled=!1)},{default:l(()=>[e(`
                Edit
              `)]),_:1}),e(),r(w,null,{default:l(()=>[e(`
                This is a unique ID for the Dataplane instance.
              `)]),_:1})])]),_:1})]),networking:l(()=>[je,e(),He,e(),Ke,e(),r(c,{"all-inline":"",title:"Data Plane IP Address","for-attr":"network-address"},{default:l(()=>[h(a("input",{id:"network-address","onUpdate:modelValue":s[6]||(s[6]=n=>t.value.univDataplaneNetworkAddress=n),type:"text",class:"k-input w-100","data-testid":"network-address"},null,512),[[N,t.value.univDataplaneNetworkAddress]]),e(),r(w,null,{default:l(()=>[e(`
              The IP address that other services will use to consume this data plane.
            `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"","fill-first":"",title:"Data Plane Port","for-attr":"network-dataplane-port"},{default:l(()=>[a("input",{id:"network-dataplane-port",value:t.value.univDataplaneNetworkDPPort,type:"text",class:"k-input w-100","data-testid":"network-dataplane-port",onInput:oe},null,40,Ye),e(),r(w,null,{default:l(()=>[e(`
              The data plane port (that other services will use to consume this service).
            `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"",title:"Service IP Address","for-attr":"network-service-address"},{default:l(()=>[h(a("input",{id:"network-service-address","onUpdate:modelValue":s[7]||(s[7]=n=>t.value.univDataplaneNetworkServiceAddress=n),type:"text",class:"k-input w-100"},null,512),[[N,t.value.univDataplaneNetworkServiceAddress]]),e(),r(w,null,{default:l(()=>[e(`
              The address where your service is listening on the machine.
            `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"",title:"Service Port","for-attr":"network-service-port"},{default:l(()=>[a("input",{id:"network-service-port",value:t.value.univDataplaneNetworkServicePort,type:"text",class:"k-input w-100","data-testid":"service-port",onInput:se},null,40,Le),e(),r(w,null,{default:l(()=>[e(`
              The port where your service is listening on the machine.
            `)]),_:1})]),_:1}),e(),r(c,{"all-inline":"",title:"Protocol","for-attr":"network-dataplane-protocol"},{default:l(()=>[h(a("select",{id:"network-dataplane-protocol","onUpdate:modelValue":s[8]||(s[8]=n=>t.value.univDataplaneNetworkProtocol=n),class:"k-input w-100",name:"network-dataplane-protocol"},[(d(!0),v(B,null,O(Q.value.protocols,(n,p)=>(d(),v("option",{key:p,value:n,selected:t.value.univDataplaneNetworkProtocol===n},m(n),9,Qe))),128))],512),[[G,t.value.univDataplaneNetworkProtocol]]),e(),r(w,null,{default:l(()=>[e(`
              The protocol of the service.
            `)]),_:1})]),_:1})]),complete:l(()=>[t.value.meshName?(d(),v("div",Ze,[x.value===!1?(d(),v("div",Xe,[Je,e(),a("p",null,`
                It's time to first generate the credentials so that `+m(u(y))+` will allow
                the Dataplane to successfully authenticate itself with the control plane,
                and then finally install the Dataplane process (powered by Envoy).
              `,1),e(),ea,e(),r(U,{id:"code-block-generate-token-command",language:"bash",code:u(X)},null,8,["code"]),e(),aa,e(),r(U,{id:"code-block-stard-dp-command",language:"bash",code:u(J)},null,8,["code"])])):W("",!0),e(),r(ke,{"loader-function":ne,"should-start":!0,"has-error":g.value,"can-complete":$.value,onHideSiblings:te},{"loading-title":l(()=>[ta]),"loading-content":l(()=>[na]),"complete-title":l(()=>[la]),"complete-content":l(()=>[a("p",null,[e(`
                  Your Dataplane
                  `),t.value.univDataplaneId?(d(),v("strong",oa,m(t.value.univDataplaneId),1)):W("",!0),e(`
                  was found!
                `)]),e(),sa,e(),a("p",null,[r(u(C),{appearance:"primary",onClick:le},{default:l(()=>[e(`
                    View Your Dataplane
                  `)]),_:1})])]),"error-title":l(()=>[ra]),"error-content":l(()=>[ia]),_:1},8,["has-error","can-complete"])])):(d(),j(u(we),{key:1,appearance:"danger"},{alertMessage:l(()=>[ua]),_:1}))]),dataplane:l(()=>[da,e(),a("p",null,`
            In `+m(u(y))+`, a Dataplane resource represents a data plane proxy running
            alongside one of your services. Data plane proxies can be added in any Mesh
            that you may have created, and in Kubernetes, they will be auto-injected
            by `+m(u(y))+`.
          `,1)]),example:l(()=>[pa,e(),ca,e(),r(U,{id:"onboarding-dpp-universal-example",class:"sample-code-block mt-3",code:P,language:"yaml"})]),switch:l(()=>[r(ge)]),_:1},8,["footer-enabled","next-disabled"])])]))}});const Na=H(va,[["__scopeId","data-v-fe363841"]]);export{Na as default};
