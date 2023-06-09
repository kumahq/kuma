import{d as R,L as Y,r as p,m as x,s as Z,$ as Q,o as u,c as S,w as n,a as i,u as r,b as a,e,t as m,P as H,U as c,Z as X,f as h,k as J,F as ee,a0 as k,V as M,g as y,p as ae,j as se}from"./index-5647cbca.js";import{Z as v,T as U,V as B}from"./kongponents.es-7466dcb0.js";import{_ as ne}from"./EntityScanner.vue_vue_type_script_setup_true_lang-6a83377e.js";import{E as te}from"./EnvironmentSwitcher-c63a1ed4.js";import{S as le,F as _}from"./StepSkeleton-4ae12edc.js";import{f as oe}from"./formatForCLI-931cd5c6.js";import{u as ie,g as re,b as de,f as ue,e as ce,_ as pe}from"./RouteView.vue_vue_type_script_setup_true_lang-f31c7713.js";import{_ as me}from"./RouteTitle.vue_vue_type_script_setup_true_lang-984571b1.js";import{_ as E}from"./CodeBlock.vue_vue_type_style_index_0_lang-898c3957.js";import{Q as he}from"./QueryParameter-70743f73.js";import"./toYaml-4e00099e.js";const ve={apiVersion:"v1",kind:"Namespace",metadata:{name:null,namespace:null,labels:{"kuma.io/sidecar-injection":"enabled"},annotations:{"kuma.io/mesh":null}}},l=f=>(ae("data-v-8d55aae9"),f=f(),se(),f),_e={class:"wizard"},ke={class:"wizard__content"},ye=l(()=>e("h3",null,`
                Create Kubernetes Dataplane
              `,-1)),ge=l(()=>e("h3",null,`
                To get started, please select on what Mesh you would like to add the Dataplane:
              `,-1)),fe=l(()=>e("p",null,`
                If you've got an existing Mesh that you would like to associate with your
                Dataplane, you can select it below, or create a new one using our Mesh Wizard.
              `,-1)),be=l(()=>e("small",null,"Would you like to see instructions for Universal? Use sidebar to change wizard!",-1)),we=l(()=>e("option",{disabled:"",value:""},`
                          Select an existing Mesh…
                        `,-1)),Se=["value"],De=l(()=>e("label",{class:"k-input-label mr-4"},`
                        or
                      `,-1)),Ie=l(()=>e("h3",null,`
                Setup Dataplane Mode
              `,-1)),Ne=l(()=>e("p",null,`
                You can create a data plane for a service or a data plane for a Gateway.
              `,-1)),xe={for:"service-dataplane"},Me=l(()=>e("span",null,`
                        Service Dataplane
                      `,-1)),Ve={for:"ingress-dataplane"},Te=l(()=>e("span",null,`
                        Ingress Dataplane
                      `,-1)),Ce={key:0},Ue=l(()=>e("p",null,`
                  Should the data plane be added for an entire Namespace and all of its services,
                  or for specific individual services in any namespace?
                `,-1)),Be={for:"k8s-services-all"},Ee=l(()=>e("span",null,`
                          All Services in Namespace
                        `,-1)),Pe={for:"k8s-services-individual"},Ke=l(()=>e("span",null,`
                          Individual Services
                        `,-1)),Fe={key:1},Ae={for:"k8s-ingress-kong"},je=l(()=>e("span",null,`
                          Kong Ingress
                        `,-1)),ze={for:"k8s-ingress-other"},$e=l(()=>e("span",null,`
                          Other Ingress
                        `,-1)),qe=l(()=>e("p",null,`
                      Please go ahead and deploy the Ingress first, then restart this wizard and select “Existing Ingress”.
                    `,-1)),Oe={key:0},We={key:0},Le=l(()=>e("h3",null,`
                    Auto-Inject DPP
                  `,-1)),Ge=l(()=>e("p",null,`
                    You can now execute the following commands to automatically inject the sidecar proxy in every Pod, and by doing so creating the Dataplane.
                  `,-1)),Re=l(()=>e("h4",null,"Kubernetes",-1)),Ye=l(()=>e("h3",null,"Searching…",-1)),Ze=l(()=>e("p",null,"We are looking for your dataplane.",-1)),Qe=l(()=>e("h3",null,"Done!",-1)),He={key:0},Xe=l(()=>e("p",null,`
                      Proceed to the next step where we will show you
                      your new Dataplane.
                    `,-1)),Je=l(()=>e("h3",null,"Mesh not found",-1)),ea=l(()=>e("p",null,"We were unable to find your mesh.",-1)),aa=l(()=>e("p",null,`
                    Please return to the first step and make sure to select an
                    existing Mesh, or create a new one.
                  `,-1)),sa=l(()=>e("h3",null,"Dataplane",-1)),na=l(()=>e("h3",null,"Example",-1)),ta=l(()=>e("p",null,`
                Below is an example of a Dataplane resource output:
              `,-1)),la=`apiVersion: 'kuma.io/v1alpha1'
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
      kuma.io/service: echo`,oa=R({__name:"DataplaneKubernetes",setup(f){const P=ie(),{t:K}=re(),F=[{label:"General",slug:"general"},{label:"Scope Settings",slug:"scope-settings"},{label:"Install",slug:"complete"}],A=[{name:"dataplane"},{name:"example"},{name:"switch"}],j=Y(),D=de(),z=p(ve),I=p(0),V=p(!1),N=p(!1),b=p(!1),T=p(!1),s=p({meshName:"",k8sDataplaneType:"dataplane-type-service",k8sServices:"all-services",k8sNamespace:"",k8sNamespaceSelection:"",k8sServiceDeployment:"",k8sServiceDeploymentSelection:"",k8sIngressDeployment:"",k8sIngressDeploymentSelection:"",k8sIngressType:"",k8sIngressBrand:"kong-ingress",k8sIngressSelection:""}),w=x(()=>D.getters["config/getTagline"]),$=x(()=>{const d=Object.assign({},z.value),t=s.value.k8sNamespaceSelection;if(!t)return"";d.metadata.name=t,d.metadata.namespace=t,d.metadata.annotations["kuma.io/mesh"]=s.value.meshName;const o=`" | kubectl apply -f - && kubectl delete pod --all -n ${t}`;return oe(d,o)}),q=x(()=>{const{k8sNamespaceSelection:d,meshName:t}=s.value;return t.length===0?!0:I.value===1?!d:!1});Z(()=>s.value.k8sNamespaceSelection,function(d){s.value.k8sNamespaceSelection=Q(d)});const C=he.get("step");I.value=C!==null?parseInt(C):0;function O(d){I.value=d}function W(){N.value=!0}async function L(){const t=s.value.meshName,o=s.value.k8sNamespaceSelection;if(T.value=!1,b.value=!1,!(!t||!o))try{const g=await P.getDataplaneFromMesh({mesh:t,name:o});g&&g.name.length>0?V.value=!0:b.value=!0}catch(g){b.value=!0,console.error(g)}finally{T.value=!0}}function G(){D.dispatch("updateSelectedMesh",s.value.meshName),j.push({name:"data-planes-list-view",params:{mesh:s.value.meshName}})}return(d,t)=>(u(),S(ce,null,{default:n(()=>[i(me,{title:r(K)("wizard-kubernetes.routes.item.title")},null,8,["title"]),a(),i(ue,null,{default:n(()=>[e("div",_e,[e("div",ke,[i(le,{steps:F,"sidebar-content":A,"footer-enabled":N.value===!1,"next-disabled":q.value,onGoToStep:O},{general:n(()=>[ye,a(),e("p",null,`
                Welcome to the wizard to create a new Dataplane resource in `+m(w.value)+`.
                We will be providing you with a few steps that will get you started.
              `,1),a(),e("p",null,`
                As you know, the `+m(r(H))+` GUI is read-only.
              `,1),a(),ge,a(),fe,a(),be,a(),i(r(v),{class:"my-6","has-shadow":""},{body:n(()=>[i(_,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:n(()=>[e("div",null,[c(e("select",{id:"dp-mesh","onUpdate:modelValue":t[0]||(t[0]=o=>s.value.meshName=o),class:"k-input w-100"},[we,a(),(u(!0),h(ee,null,J(r(D).getters.getMeshList.items,o=>(u(),h("option",{key:o.name,value:o.name},m(o.name),9,Se))),128))],512),[[X,s.value.meshName]])]),a(),e("div",null,[De,a(),i(r(U),{to:{name:"create-mesh"},appearance:"outline"},{default:n(()=>[a(`
                        Create a new Mesh
                      `)]),_:1})])]),_:1})]),_:1})]),"scope-settings":n(()=>[Ie,a(),Ne,a(),i(r(v),{class:"my-6","has-shadow":""},{body:n(()=>[i(_,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:n(()=>[e("label",xe,[c(e("input",{id:"service-dataplane","onUpdate:modelValue":t[1]||(t[1]=o=>s.value.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[k,s.value.k8sDataplaneType]]),a(),Me]),a(),e("label",Ve,[c(e("input",{id:"ingress-dataplane","onUpdate:modelValue":t[2]||(t[2]=o=>s.value.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-ingress",disabled:""},null,512),[[k,s.value.k8sDataplaneType]]),a(),Te])]),_:1})]),_:1}),a(),s.value.k8sDataplaneType==="dataplane-type-service"?(u(),h("div",Ce,[Ue,a(),i(r(v),{class:"my-6","has-shadow":""},{body:n(()=>[i(_,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:n(()=>[e("label",Be,[c(e("input",{id:"k8s-services-all","onUpdate:modelValue":t[3]||(t[3]=o=>s.value.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"all-services",checked:""},null,512),[[k,s.value.k8sServices]]),a(),Ee]),a(),e("label",Pe,[c(e("input",{id:"k8s-services-individual","onUpdate:modelValue":t[4]||(t[4]=o=>s.value.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"individual-services",disabled:""},null,512),[[k,s.value.k8sServices]]),a(),Ke])]),_:1})]),_:1}),a(),s.value.k8sServices==="individual-services"?(u(),S(r(v),{key:0,class:"my-6","has-shadow":""},{body:n(()=>[i(_,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:n(()=>[c(e("input",{id:"k8s-service-deployment-new","onUpdate:modelValue":t[5]||(t[5]=o=>s.value.k8sServiceDeploymentSelection=o),type:"text",class:"k-input w-100",placeholder:"your-new-deployment",required:""},null,512),[[M,s.value.k8sServiceDeploymentSelection]])]),_:1})]),_:1})):y("",!0),a(),i(r(v),{class:"my-6","has-shadow":""},{body:n(()=>[i(_,{title:"Namespace","for-attr":"k8s-namespace-selection"},{default:n(()=>[c(e("input",{id:"k8s-namespace-new","onUpdate:modelValue":t[6]||(t[6]=o=>s.value.k8sNamespaceSelection=o),type:"text",class:"k-input w-100",placeholder:"your-namespace",required:""},null,512),[[M,s.value.k8sNamespaceSelection]])]),_:1})]),_:1})])):y("",!0),a(),s.value.k8sDataplaneType==="dataplane-type-ingress"?(u(),h("div",Fe,[e("p",null,m(w.value)+` natively supports the Kong Ingress. Do you want to deploy
                  Kong or another Ingress?
                `,1),a(),i(r(v),{class:"my-6","has-shadow":""},{body:n(()=>[i(_,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:n(()=>[e("label",Ae,[c(e("input",{id:"k8s-ingress-kong","onUpdate:modelValue":t[7]||(t[7]=o=>s.value.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"kong-ingress",checked:""},null,512),[[k,s.value.k8sIngressBrand]]),a(),je]),a(),e("label",ze,[c(e("input",{id:"k8s-ingress-other","onUpdate:modelValue":t[8]||(t[8]=o=>s.value.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"other-ingress"},null,512),[[k,s.value.k8sIngressBrand]]),a(),$e])]),_:1})]),_:1}),a(),i(r(v),{class:"my-6","has-shadow":""},{body:n(()=>[i(_,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:n(()=>[c(e("input",{id:"k8s-ingress-deployment-new","onUpdate:modelValue":t[9]||(t[9]=o=>s.value.k8sIngressDeployment=o),type:"text",class:"k-input w-100",placeholder:"your-deployment",required:""},null,512),[[M,s.value.k8sIngressDeployment]])]),_:1})]),_:1}),a(),s.value.k8sIngressBrand==="other-ingress"?(u(),S(r(B),{key:0,appearance:"info"},{alertMessage:n(()=>[qe]),_:1})):y("",!0)])):y("",!0)]),complete:n(()=>[s.value.meshName?(u(),h("div",Oe,[N.value===!1?(u(),h("div",We,[Le,a(),Ge,a(),Re,a(),i(E,{id:"code-block-kubernetes-command",class:"mt-3",language:"bash",code:$.value},null,8,["code"])])):y("",!0),a(),i(ne,{"loader-function":L,"has-error":b.value,"can-complete":V.value,onHideSiblings:W},{"loading-title":n(()=>[Ye]),"loading-content":n(()=>[Ze]),"complete-title":n(()=>[Qe]),"complete-content":n(()=>[e("p",null,[a(`
                      Your Dataplane
                      `),s.value.k8sNamespaceSelection?(u(),h("strong",He,m(s.value.k8sNamespaceSelection),1)):y("",!0),a(`
                      was found!
                    `)]),a(),Xe,a(),e("p",null,[i(r(U),{appearance:"primary",onClick:G},{default:n(()=>[a(`
                        View Your Dataplane
                      `)]),_:1})])]),"error-title":n(()=>[Je]),"error-content":n(()=>[ea]),_:1},8,["has-error","can-complete"])])):(u(),S(r(B),{key:1,appearance:"danger"},{alertMessage:n(()=>[aa]),_:1}))]),dataplane:n(()=>[sa,a(),e("p",null,`
                In `+m(w.value)+`, a Dataplane resource represents a data plane proxy running
                alongside one of your services. Data plane proxies can be added in any Mesh
                that you may have created, and in Kubernetes, they will be auto-injected
                by `+m(w.value)+`.
              `,1)]),example:n(()=>[na,a(),ta,a(),i(E,{id:"onboarding-dpp-kubernetes-example",class:"sample-code-block",code:la,language:"yaml"})]),switch:n(()=>[i(te)]),_:1},8,["footer-enabled","next-disabled"])])])]),_:1})]),_:1}))}});const ya=pe(oa,[["__scopeId","data-v-8d55aae9"]]);export{ya as default};
