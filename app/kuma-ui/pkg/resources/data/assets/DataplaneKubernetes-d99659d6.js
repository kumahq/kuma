import{d as R,a as Y,r as m,c as N,v as $,a1 as Q,o as u,j as p,i as e,g as i,w as t,h as a,t as h,e as r,P as X,W as c,$ as Z,F as H,q as J,a2 as k,b as x,X as M,f as y,p as ee,m as ae}from"./index-b62490ef.js";import{z as v,E as C,Z as B}from"./kongponents.es-7d82be84.js";import{_ as se}from"./EntityScanner.vue_vue_type_script_setup_true_lang-c44c4156.js";import{E as ne}from"./EnvironmentSwitcher-ec0f98c4.js";import{F as _,S as te}from"./StepSkeleton-00ba8e6d.js";import{f as le}from"./formatForCLI-931cd5c6.js";import{_ as P}from"./CodeBlock.vue_vue_type_style_index_0_lang-357021b0.js";import{u as oe}from"./store-46e112fa.js";import{u as ie}from"./index-8249b827.js";import{Q as re}from"./QueryParameter-70743f73.js";import{_ as de}from"./_plugin-vue_export-helper-c27b6911.js";import"./toYaml-4e00099e.js";const ue={apiVersion:"v1",kind:"Namespace",metadata:{name:null,namespace:null,labels:{"kuma.io/sidecar-injection":"enabled"},annotations:{"kuma.io/mesh":null}}},l=f=>(ee("data-v-65556661"),f=f(),ae(),f),ce={class:"wizard"},pe={class:"wizard__content"},me=l(()=>e("h3",null,`
            Create Kubernetes Dataplane
          `,-1)),he=l(()=>e("h3",null,`
            To get started, please select on what Mesh you would like to add the Dataplane:
          `,-1)),ve=l(()=>e("p",null,`
            If you've got an existing Mesh that you would like to associate with your
            Dataplane, you can select it below, or create a new one using our Mesh Wizard.
          `,-1)),_e=l(()=>e("small",null,"Would you like to see instructions for Universal? Use sidebar to change wizard!",-1)),ke=l(()=>e("option",{disabled:"",value:""},`
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
                    `,-1)),Ce={key:1},Be={for:"k8s-ingress-kong"},Pe=l(()=>e("span",null,`
                      Kong Ingress
                    `,-1)),Ue={for:"k8s-ingress-other"},Ke=l(()=>e("span",null,`
                      Other Ingress
                    `,-1)),Fe=l(()=>e("p",null,`
                  Please go ahead and deploy the Ingress first, then restart this wizard and select “Existing Ingress”.
                `,-1)),ze={key:0},Ae={key:0},je=l(()=>e("h3",null,`
                Auto-Inject DPP
              `,-1)),qe=l(()=>e("p",null,`
                You can now execute the following commands to automatically inject the sidecar proxy in every Pod, and by doing so creating the Dataplane.
              `,-1)),We=l(()=>e("h4",null,"Kubernetes",-1)),Oe=l(()=>e("h3",null,"Searching…",-1)),Ge=l(()=>e("p",null,"We are looking for your dataplane.",-1)),Le=l(()=>e("h3",null,"Done!",-1)),Re={key:0},Ye=l(()=>e("p",null,`
                  Proceed to the next step where we will show you
                  your new Dataplane.
                `,-1)),$e=l(()=>e("h3",null,"Mesh not found",-1)),Qe=l(()=>e("p",null,"We were unable to find your mesh.",-1)),Xe=l(()=>e("p",null,`
                Please return to the first step and make sure to select an
                existing Mesh, or create a new one.
              `,-1)),Ze=l(()=>e("h3",null,"Dataplane",-1)),He=l(()=>e("h3",null,"Example",-1)),Je=l(()=>e("p",null,`
            Below is an example of a Dataplane resource output:
          `,-1)),ea=`apiVersion: 'kuma.io/v1alpha1'
kind: Dataplane
mesh: default
metadata:
  name: dp-echo-1
  labels:
    kuma.io/sidecar-injection: enabled
  annotations:
    kuma.io/mesh: default
networking:
  address: 10.0.0.1
  inbound:
  - port: 10000
    servicePort: 9000
    tags:
      kuma.io/service: echo`,aa=R({__name:"DataplaneKubernetes",setup(f){const U=ie(),K=[{label:"General",slug:"general"},{label:"Scope Settings",slug:"scope-settings"},{label:"Install",slug:"complete"}],F=[{name:"dataplane"},{name:"example"},{name:"switch"}],z=Y(),S=oe(),A=m(ue),D=m(0),T=m(!1),I=m(!1),b=m(!1),V=m(!1),s=m({meshName:"",k8sDataplaneType:"dataplane-type-service",k8sServices:"all-services",k8sNamespace:"",k8sNamespaceSelection:"",k8sServiceDeployment:"",k8sServiceDeploymentSelection:"",k8sIngressDeployment:"",k8sIngressDeploymentSelection:"",k8sIngressType:"",k8sIngressBrand:"kong-ingress",k8sIngressSelection:""}),w=N(()=>S.getters["config/getTagline"]),j=N(()=>{const d=Object.assign({},A.value),n=s.value.k8sNamespaceSelection;if(!n)return"";d.metadata.name=n,d.metadata.namespace=n,d.metadata.annotations["kuma.io/mesh"]=s.value.meshName;const o=`" | kubectl apply -f - && kubectl delete pod --all -n ${n}`;return le(d,o)}),q=N(()=>{const{k8sNamespaceSelection:d,meshName:n}=s.value;return n.length===0?!0:D.value===1?!d:!1});$(()=>s.value.k8sNamespaceSelection,function(d){s.value.k8sNamespaceSelection=Q(d)});const E=re.get("step");D.value=E!==null?parseInt(E):0;function W(d){D.value=d}function O(){I.value=!0}async function G(){const n=s.value.meshName,o=s.value.k8sNamespaceSelection;if(V.value=!1,b.value=!1,!(!n||!o))try{const g=await U.getDataplaneFromMesh({mesh:n,name:o});g&&g.name.length>0?T.value=!0:b.value=!0}catch(g){b.value=!0,console.error(g)}finally{V.value=!0}}function L(){S.dispatch("updateSelectedMesh",s.value.meshName),z.push({name:"data-planes-list-view",params:{mesh:s.value.meshName}})}return(d,n)=>(u(),p("div",ce,[e("div",pe,[i(te,{steps:K,"sidebar-content":F,"footer-enabled":I.value===!1,"next-disabled":q.value,onGoToStep:W},{general:t(()=>[me,a(),e("p",null,`
            Welcome to the wizard to create a new Dataplane resource in `+h(w.value)+`.
            We will be providing you with a few steps that will get you started.
          `,1),a(),e("p",null,`
            As you know, the `+h(r(X))+` GUI is read-only.
          `,1),a(),he,a(),ve,a(),_e,a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(_,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:t(()=>[e("div",null,[c(e("select",{id:"dp-mesh","onUpdate:modelValue":n[0]||(n[0]=o=>s.value.meshName=o),class:"k-input w-100"},[ke,a(),(u(!0),p(H,null,J(r(S).getters.getMeshList.items,o=>(u(),p("option",{key:o.name,value:o.name},h(o.name),9,ye))),128))],512),[[Z,s.value.meshName]])]),a(),e("div",null,[ge,a(),i(r(C),{to:{name:"create-mesh"},appearance:"outline"},{default:t(()=>[a(`
                    Create a new Mesh
                  `)]),_:1})])]),_:1})]),_:1})]),"scope-settings":t(()=>[fe,a(),be,a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(_,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:t(()=>[e("label",we,[c(e("input",{id:"service-dataplane","onUpdate:modelValue":n[1]||(n[1]=o=>s.value.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[k,s.value.k8sDataplaneType]]),a(),Se]),a(),e("label",De,[c(e("input",{id:"ingress-dataplane","onUpdate:modelValue":n[2]||(n[2]=o=>s.value.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-ingress",disabled:""},null,512),[[k,s.value.k8sDataplaneType]]),a(),Ie])]),_:1})]),_:1}),a(),s.value.k8sDataplaneType==="dataplane-type-service"?(u(),p("div",Ne,[xe,a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(_,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:t(()=>[e("label",Me,[c(e("input",{id:"k8s-services-all","onUpdate:modelValue":n[3]||(n[3]=o=>s.value.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"all-services",checked:""},null,512),[[k,s.value.k8sServices]]),a(),Te]),a(),e("label",Ve,[c(e("input",{id:"k8s-services-individual","onUpdate:modelValue":n[4]||(n[4]=o=>s.value.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"individual-services",disabled:""},null,512),[[k,s.value.k8sServices]]),a(),Ee])]),_:1})]),_:1}),a(),s.value.k8sServices==="individual-services"?(u(),x(r(v),{key:0,class:"my-6","has-shadow":""},{body:t(()=>[i(_,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:t(()=>[c(e("input",{id:"k8s-service-deployment-new","onUpdate:modelValue":n[5]||(n[5]=o=>s.value.k8sServiceDeploymentSelection=o),type:"text",class:"k-input w-100",placeholder:"your-new-deployment",required:""},null,512),[[M,s.value.k8sServiceDeploymentSelection]])]),_:1})]),_:1})):y("",!0),a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(_,{title:"Namespace","for-attr":"k8s-namespace-selection"},{default:t(()=>[c(e("input",{id:"k8s-namespace-new","onUpdate:modelValue":n[6]||(n[6]=o=>s.value.k8sNamespaceSelection=o),type:"text",class:"k-input w-100",placeholder:"your-namespace",required:""},null,512),[[M,s.value.k8sNamespaceSelection]])]),_:1})]),_:1})])):y("",!0),a(),s.value.k8sDataplaneType==="dataplane-type-ingress"?(u(),p("div",Ce,[e("p",null,h(w.value)+` natively supports the Kong Ingress. Do you want to deploy
              Kong or another Ingress?
            `,1),a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(_,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:t(()=>[e("label",Be,[c(e("input",{id:"k8s-ingress-kong","onUpdate:modelValue":n[7]||(n[7]=o=>s.value.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"kong-ingress",checked:""},null,512),[[k,s.value.k8sIngressBrand]]),a(),Pe]),a(),e("label",Ue,[c(e("input",{id:"k8s-ingress-other","onUpdate:modelValue":n[8]||(n[8]=o=>s.value.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"other-ingress"},null,512),[[k,s.value.k8sIngressBrand]]),a(),Ke])]),_:1})]),_:1}),a(),i(r(v),{class:"my-6","has-shadow":""},{body:t(()=>[i(_,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:t(()=>[c(e("input",{id:"k8s-ingress-deployment-new","onUpdate:modelValue":n[9]||(n[9]=o=>s.value.k8sIngressDeployment=o),type:"text",class:"k-input w-100",placeholder:"your-deployment",required:""},null,512),[[M,s.value.k8sIngressDeployment]])]),_:1})]),_:1}),a(),s.value.k8sIngressBrand==="other-ingress"?(u(),x(r(B),{key:0,appearance:"info"},{alertMessage:t(()=>[Fe]),_:1})):y("",!0)])):y("",!0)]),complete:t(()=>[s.value.meshName?(u(),p("div",ze,[I.value===!1?(u(),p("div",Ae,[je,a(),qe,a(),We,a(),i(P,{id:"code-block-kubernetes-command",class:"mt-3",language:"bash",code:j.value},null,8,["code"])])):y("",!0),a(),i(se,{"loader-function":G,"has-error":b.value,"can-complete":T.value,onHideSiblings:O},{"loading-title":t(()=>[Oe]),"loading-content":t(()=>[Ge]),"complete-title":t(()=>[Le]),"complete-content":t(()=>[e("p",null,[a(`
                  Your Dataplane
                  `),s.value.k8sNamespaceSelection?(u(),p("strong",Re,h(s.value.k8sNamespaceSelection),1)):y("",!0),a(`
                  was found!
                `)]),a(),Ye,a(),e("p",null,[i(r(C),{appearance:"primary",onClick:L},{default:t(()=>[a(`
                    View Your Dataplane
                  `)]),_:1})])]),"error-title":t(()=>[$e]),"error-content":t(()=>[Qe]),_:1},8,["has-error","can-complete"])])):(u(),x(r(B),{key:1,appearance:"danger"},{alertMessage:t(()=>[Xe]),_:1}))]),dataplane:t(()=>[Ze,a(),e("p",null,`
            In `+h(w.value)+`, a Dataplane resource represents a data plane proxy running
            alongside one of your services. Data plane proxies can be added in any Mesh
            that you may have created, and in Kubernetes, they will be auto-injected
            by `+h(w.value)+`.
          `,1)]),example:t(()=>[He,a(),Je,a(),i(P,{id:"onboarding-dpp-kubernetes-example",class:"sample-code-block",code:ea,language:"yaml"})]),switch:t(()=>[i(ne)]),_:1},8,["footer-enabled","next-disabled"])])]))}});const ha=de(aa,[["__scopeId","data-v-65556661"]]);export{ha as default};
