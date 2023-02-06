import{d as R,r as m,c as N,s as Y,h as p,g as e,e as i,w as t,u as r,o as u,f as a,t as h,S as c,a6 as $,F as Q,m as H,a7 as _,a as x,a5 as M,b as y,p as X,j as J}from"./runtime-dom.esm-bundler-32659b48.js";import{b as Z}from"./vue-router-d8e03a07.js";import{$ as v,P,T as C}from"./kongponents.es-c2485d1e.js";import{f as ee}from"./formatForCLI-931cd5c6.js";import{D as ae,P as se}from"./production-c33f040b.js";import{k as ne}from"./kumaApi-302acf12.js";import{Q as te}from"./QueryParameter-70743f73.js";import{u as le}from"./store-96085224.js";import{_ as B}from"./CodeBlock.vue_vue_type_style_index_0_lang-a6bd5c2b.js";import{S as oe,F as k,E as ie}from"./StepSkeleton-fa974c6d.js";import{E as re}from"./EnvironmentSwitcher-b5150e63.js";import{_ as de}from"./_plugin-vue_export-helper-c27b6911.js";import"./toYaml-4e00099e.js";import"./_commonjsHelpers-edff4021.js";const ue={apiVersion:"v1",kind:"Namespace",metadata:{name:null,namespace:null,annotations:{"kuma.io/sidecar-injection":"enabled","kuma.io/mesh":null}}},l=f=>(X("data-v-677644d2"),f=f(),J(),f),ce={class:"wizard"},pe={class:"wizard__content"},me=l(()=>e("h3",null,`
            Create Kubernetes Dataplane
          `,-1)),he=l(()=>e("h3",null,`
            To get started, please select on what Mesh you would like to add the Dataplane:
          `,-1)),ve=l(()=>e("p",null,`
            If you've got an existing Mesh that you would like to associate with your
            Dataplane, you can select it below, or create a new one using our Mesh Wizard.
          `,-1)),ke=l(()=>e("small",null,"Would you like to see instructions for Universal? Use sidebar to change wizard!",-1)),_e=l(()=>e("option",{disabled:"",value:""},`
                      Select an existing Mesh…
                    `,-1)),ye=["value"],ge=l(()=>e("label",{class:"k-input-label mr-4"},`
                    or
                  `,-1)),fe=l(()=>e("h3",null,`
            Setup Dataplane Mode
          `,-1)),be=l(()=>e("p",null,`
            You can create a data plane for a service or a data plane for a Gateway.
          `,-1)),we={for:"service-dataplane"},Se=l(()=>e("span",null,`
                    Service Dataplane
                  `,-1)),De={for:"ingress-dataplane"},Ie=l(()=>e("span",null,`
                    Ingress Dataplane
                  `,-1)),Ne={key:0},xe=l(()=>e("p",null,`
              Should the data plane be added for an entire Namespace and all of its services,
              or for specific individual services in any namespace?
            `,-1)),Me={for:"k8s-services-all"},Te=l(()=>e("span",null,`
                      All Services in Namespace
                    `,-1)),Ve={for:"k8s-services-individual"},Ee=l(()=>e("span",null,`
                      Individual Services
                    `,-1)),Pe={key:1},Ce={for:"k8s-ingress-kong"},Be=l(()=>e("span",null,`
                      Kong Ingress
                    `,-1)),Ue={for:"k8s-ingress-other"},Fe=l(()=>e("span",null,`
                      Other Ingress
                    `,-1)),Ke=l(()=>e("p",null,`
                  Please go ahead and deploy the Ingress first, then restart this wizard and select “Existing Ingress”.
                `,-1)),je={key:0},Ae={key:0},qe=l(()=>e("h3",null,`
                Auto-Inject DPP
              `,-1)),ze=l(()=>e("p",null,`
                You can now execute the following commands to automatically inject the sidecar proxy in every Pod, and by doing so creating the Dataplane.
              `,-1)),Oe=l(()=>e("h4",null,"Kubernetes",-1)),We=l(()=>e("h3",null,"Searching…",-1)),Ge=l(()=>e("p",null,"We are looking for your dataplane.",-1)),Le=l(()=>e("h3",null,"Done!",-1)),Re={key:0},Ye=l(()=>e("p",null,`
                  Proceed to the next step where we will show you
                  your new Dataplane.
                `,-1)),$e=l(()=>e("h3",null,"Mesh not found",-1)),Qe=l(()=>e("p",null,"We were unable to find your mesh.",-1)),He=l(()=>e("p",null,`
                Please return to the first step and make sure to select an
                existing Mesh, or create a new one.
              `,-1)),Xe=l(()=>e("h3",null,"Dataplane",-1)),Je=l(()=>e("h3",null,"Example",-1)),Ze=l(()=>e("p",null,`
            Below is an example of a Dataplane resource output:
          `,-1)),ea=R({__name:"DataplaneKubernetes",setup(f){const U=`apiVersion: 'kuma.io/v1alpha1'
kind: Dataplane
mesh: default
metadata:
  name: dp-echo-1
  annotations:
    kuma.io/sidecar-injection: enabled
    kuma.io/mesh: default
networking:
  address: 10.0.0.1
  inbound:
  - port: 10000
    servicePort: 9000
    tags:
      kuma.io/service: echo`,F=[{label:"General",slug:"general"},{label:"Scope Settings",slug:"scope-settings"},{label:"Install",slug:"complete"}],K=[{name:"dataplane"},{name:"example"},{name:"switch"}],j=Z(),S=le(),A=m(ue),D=m(0),T=m(!1),I=m(!1),b=m(!1),V=m(!1),s=m({meshName:"",k8sDataplaneType:"dataplane-type-service",k8sServices:"all-services",k8sNamespace:"",k8sNamespaceSelection:"",k8sServiceDeployment:"",k8sServiceDeploymentSelection:"",k8sIngressDeployment:"",k8sIngressDeploymentSelection:"",k8sIngressType:"",k8sIngressBrand:"kong-ingress",k8sIngressSelection:""}),w=N(()=>S.getters["config/getTagline"]),q=N(()=>{const d=Object.assign({},A.value),n=s.value.k8sNamespaceSelection;if(!n)return"";d.metadata.name=n,d.metadata.namespace=n,d.metadata.annotations["kuma.io/mesh"]=s.value.meshName;const o=`" | kubectl apply -f - && kubectl delete pod --all -n ${n}`;return ee(d,o)}),z=N(()=>{const{k8sNamespaceSelection:d,meshName:n}=s.value;return n.length===0?!0:D.value===1?!d:!1});Y(()=>s.value.k8sNamespaceSelection,function(d){s.value.k8sNamespaceSelection=ae(d)});const E=te.get("step");D.value=E!==null?parseInt(E):0;function O(d){D.value=d}function W(){I.value=!0}async function G(){const n=s.value.meshName,o=s.value.k8sNamespaceSelection;if(V.value=!1,b.value=!1,!(!n||!o))try{const g=await ne.getDataplaneFromMesh({mesh:n,name:o});g&&g.name.length>0?T.value=!0:b.value=!0}catch(g){b.value=!0,console.error(g)}finally{V.value=!0}}function L(){S.dispatch("updateSelectedMesh",s.value.meshName),j.push({name:"data-plane-list-view",params:{mesh:s.value.meshName}})}return(d,n)=>(u(),p("div",ce,[e("div",pe,[i(oe,{steps:F,"sidebar-content":K,"footer-enabled":I.value===!1,"next-disabled":r(z),onGoToStep:O},{general:t(()=>[me,a(),e("p",null,`
            Welcome to the wizard to create a new Dataplane resource in `+h(r(w))+`.
            We will be providing you with a few steps that will get you started.
          `,1),a(),e("p",null,`
            As you know, the `+h(r(se))+` GUI is read-only.
          `,1),a(),he,a(),ve,a(),ke,a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(k,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:t(()=>[e("div",null,[c(e("select",{id:"dp-mesh","onUpdate:modelValue":n[0]||(n[0]=o=>s.value.meshName=o),class:"k-input w-100"},[_e,a(),(u(!0),p(Q,null,H(r(S).getters.getMeshList.items,o=>(u(),p("option",{key:o.name,value:o.name},h(o.name),9,ye))),128))],512),[[$,s.value.meshName]])]),a(),e("div",null,[ge,a(),i(r(P),{to:{name:"create-mesh"},appearance:"outline"},{default:t(()=>[a(`
                    Create a new Mesh
                  `)]),_:1})])]),_:1})]),_:1})]),"scope-settings":t(()=>[fe,a(),be,a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(k,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:t(()=>[e("label",we,[c(e("input",{id:"service-dataplane","onUpdate:modelValue":n[1]||(n[1]=o=>s.value.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[_,s.value.k8sDataplaneType]]),a(),Se]),a(),e("label",De,[c(e("input",{id:"ingress-dataplane","onUpdate:modelValue":n[2]||(n[2]=o=>s.value.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-ingress",disabled:""},null,512),[[_,s.value.k8sDataplaneType]]),a(),Ie])]),_:1})]),_:1}),a(),s.value.k8sDataplaneType==="dataplane-type-service"?(u(),p("div",Ne,[xe,a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(k,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:t(()=>[e("label",Me,[c(e("input",{id:"k8s-services-all","onUpdate:modelValue":n[3]||(n[3]=o=>s.value.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"all-services",checked:""},null,512),[[_,s.value.k8sServices]]),a(),Te]),a(),e("label",Ve,[c(e("input",{id:"k8s-services-individual","onUpdate:modelValue":n[4]||(n[4]=o=>s.value.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"individual-services",disabled:""},null,512),[[_,s.value.k8sServices]]),a(),Ee])]),_:1})]),_:1}),a(),s.value.k8sServices==="individual-services"?(u(),x(r(v),{key:0,class:"my-6","has-shadow":""},{body:t(()=>[i(k,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:t(()=>[c(e("input",{id:"k8s-service-deployment-new","onUpdate:modelValue":n[5]||(n[5]=o=>s.value.k8sServiceDeploymentSelection=o),type:"text",class:"k-input w-100",placeholder:"your-new-deployment",required:""},null,512),[[M,s.value.k8sServiceDeploymentSelection]])]),_:1})]),_:1})):y("",!0),a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(k,{title:"Namespace","for-attr":"k8s-namespace-selection"},{default:t(()=>[c(e("input",{id:"k8s-namespace-new","onUpdate:modelValue":n[6]||(n[6]=o=>s.value.k8sNamespaceSelection=o),type:"text",class:"k-input w-100",placeholder:"your-namespace",required:""},null,512),[[M,s.value.k8sNamespaceSelection]])]),_:1})]),_:1})])):y("",!0),a(),s.value.k8sDataplaneType==="dataplane-type-ingress"?(u(),p("div",Pe,[e("p",null,h(r(w))+` natively supports the Kong Ingress. Do you want to deploy
              Kong or another Ingress?
            `,1),a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(k,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:t(()=>[e("label",Ce,[c(e("input",{id:"k8s-ingress-kong","onUpdate:modelValue":n[7]||(n[7]=o=>s.value.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"kong-ingress",checked:""},null,512),[[_,s.value.k8sIngressBrand]]),a(),Be]),a(),e("label",Ue,[c(e("input",{id:"k8s-ingress-other","onUpdate:modelValue":n[8]||(n[8]=o=>s.value.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"other-ingress"},null,512),[[_,s.value.k8sIngressBrand]]),a(),Fe])]),_:1})]),_:1}),a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(k,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:t(()=>[c(e("input",{id:"k8s-ingress-deployment-new","onUpdate:modelValue":n[9]||(n[9]=o=>s.value.k8sIngressDeployment=o),type:"text",class:"k-input w-100",placeholder:"your-deployment",required:""},null,512),[[M,s.value.k8sIngressDeployment]])]),_:1})]),_:1}),a(),s.value.k8sIngressBrand==="other-ingress"?(u(),x(r(C),{key:0,appearance:"info"},{alertMessage:t(()=>[Ke]),_:1})):y("",!0)])):y("",!0)]),complete:t(()=>[s.value.meshName?(u(),p("div",je,[I.value===!1?(u(),p("div",Ae,[qe,a(),ze,a(),Oe,a(),i(B,{id:"code-block-kubernetes-command",class:"mt-3",language:"bash",code:r(q)},null,8,["code"])])):y("",!0),a(),i(ie,{"loader-function":G,"should-start":!0,"has-error":b.value,"can-complete":T.value,onHideSiblings:W},{"loading-title":t(()=>[We]),"loading-content":t(()=>[Ge]),"complete-title":t(()=>[Le]),"complete-content":t(()=>[e("p",null,[a(`
                  Your Dataplane
                  `),s.value.k8sNamespaceSelection?(u(),p("strong",Re,h(s.value.k8sNamespaceSelection),1)):y("",!0),a(`
                  was found!
                `)]),a(),Ye,a(),e("p",null,[i(r(P),{appearance:"primary",onClick:L},{default:t(()=>[a(`
                    View Your Dataplane
                  `)]),_:1})])]),"error-title":t(()=>[$e]),"error-content":t(()=>[Qe]),_:1},8,["has-error","can-complete"])])):(u(),x(r(C),{key:1,appearance:"danger"},{alertMessage:t(()=>[He]),_:1}))]),dataplane:t(()=>[Xe,a(),e("p",null,`
            In `+h(r(w))+`, a Dataplane resource represents a data plane proxy running
            alongside one of your services. Data plane proxies can be added in any Mesh
            that you may have created, and in Kubernetes, they will be auto-injected
            by `+h(r(w))+`.
          `,1)]),example:t(()=>[Je,a(),Ze,a(),i(B,{id:"onboarding-dpp-kubernetes-example",class:"sample-code-block",code:U,language:"yaml"})]),switch:t(()=>[i(re)]),_:1},8,["footer-enabled","next-disabled"])])]))}});const va=de(ea,[["__scopeId","data-v-677644d2"]]);export{va as default};
