import{d as L,v as Y,q as p,c as U,s as R,o as u,a as S,w as s,h as i,b as r,g as a,k as e,t as m,D as c,G as Q,e as h,j as Z,F as H,J as y,E as N,f as g,p as J,m as X}from"./index-52119cf8.js";import{U as v,S as B,Z as C}from"./kongponents.es-6866aec3.js";import{_ as ee}from"./EntityScanner.vue_vue_type_script_setup_true_lang-de50e8e8.js";import{E as ae}from"./EnvironmentSwitcher-9057f08d.js";import{S as ne,F as _}from"./StepSkeleton-4bf97892.js";import{f as se}from"./formatForCLI-931cd5c6.js";import{j as te,k as le,f as oe,K as ie,g as re,_ as de,h as ue}from"./RouteView.vue_vue_type_script_setup_true_lang-28e29218.js";import{_ as ce}from"./RouteTitle.vue_vue_type_script_setup_true_lang-a8ac9aaf.js";import{_ as E}from"./CodeBlock.vue_vue_type_style_index_0_lang-3c985a9b.js";import{Q as pe}from"./QueryParameter-70743f73.js";import"./toYaml-4e00099e.js";const me={apiVersion:"v1",kind:"Namespace",metadata:{name:null,namespace:null,labels:{"kuma.io/sidecar-injection":"enabled"},annotations:{"kuma.io/mesh":null}}},l=b=>(J("data-v-6e5d7460"),b=b(),X(),b),he={class:"wizard"},ve={class:"wizard__content"},_e=l(()=>e("h3",null,`
                Create Kubernetes Dataplane
              `,-1)),ke=l(()=>e("h3",null,`
                To get started, please select on what Mesh you would like to add the Dataplane:
              `,-1)),ye=l(()=>e("p",null,`
                If you've got an existing Mesh that you would like to associate with your
                Dataplane, you can select it below, or create a new one using our Mesh Wizard.
              `,-1)),ge=l(()=>e("small",null,"Would you like to see instructions for Universal? Use sidebar to change wizard!",-1)),fe=l(()=>e("option",{disabled:"",value:""},`
                          Select an existing Mesh…
                        `,-1)),be=["value"],we=l(()=>e("label",{class:"k-input-label mr-4"},`
                        or
                      `,-1)),Se=l(()=>e("h3",null,`
                Setup Dataplane Mode
              `,-1)),De=l(()=>e("p",null,`
                You can create a data plane for a service or a data plane for a Gateway.
              `,-1)),Ie={for:"service-dataplane"},Ne=l(()=>e("span",null,`
                        Service Dataplane
                      `,-1)),xe={for:"ingress-dataplane"},Me=l(()=>e("span",null,`
                        Ingress Dataplane
                      `,-1)),Ve={key:0},Te=l(()=>e("p",null,`
                  Should the data plane be added for an entire Namespace and all of its services,
                  or for specific individual services in any namespace?
                `,-1)),Ue={for:"k8s-services-all"},Be=l(()=>e("span",null,`
                          All Services in Namespace
                        `,-1)),Ce={for:"k8s-services-individual"},Ee=l(()=>e("span",null,`
                          Individual Services
                        `,-1)),Pe={key:1},Ke={for:"k8s-ingress-kong"},Fe=l(()=>e("span",null,`
                          Kong Ingress
                        `,-1)),je={for:"k8s-ingress-other"},qe=l(()=>e("span",null,`
                          Other Ingress
                        `,-1)),ze=l(()=>e("p",null,`
                      Please go ahead and deploy the Ingress first, then restart this wizard and select “Existing Ingress”.
                    `,-1)),Ae={key:0},We={key:0},$e=l(()=>e("h3",null,`
                    Auto-Inject DPP
                  `,-1)),Ge=l(()=>e("p",null,`
                    You can now execute the following commands to automatically inject the sidecar proxy in every Pod, and by doing so creating the Dataplane.
                  `,-1)),Oe=l(()=>e("h4",null,"Kubernetes",-1)),Le=l(()=>e("h3",null,"Searching…",-1)),Ye=l(()=>e("p",null,"We are looking for your dataplane.",-1)),Re=l(()=>e("h3",null,"Done!",-1)),Qe={key:0},Ze=l(()=>e("p",null,`
                      Proceed to the next step where we will show you
                      your new Dataplane.
                    `,-1)),He=l(()=>e("h3",null,"Mesh not found",-1)),Je=l(()=>e("p",null,"We were unable to find your mesh.",-1)),Xe=l(()=>e("p",null,`
                    Please return to the first step and make sure to select an
                    existing Mesh, or create a new one.
                  `,-1)),ea=l(()=>e("h3",null,"Dataplane",-1)),aa=l(()=>e("h3",null,"Example",-1)),na=l(()=>e("p",null,`
                Below is an example of a Dataplane resource output:
              `,-1)),sa=`apiVersion: 'kuma.io/v1alpha1'
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
      kuma.io/service: echo`,ta=L({__name:"DataplaneKubernetes",setup(b){const P=te(),{t:k}=le(),K=[{label:"General",slug:"general"},{label:"Scope Settings",slug:"scope-settings"},{label:"Install",slug:"complete"}],F=[{name:"dataplane"},{name:"example"},{name:"switch"}],j=Y(),x=oe(),q=p(me),D=p(0),M=p(!1),I=p(!1),w=p(!1),V=p(!1),n=p({meshName:"",k8sDataplaneType:"dataplane-type-service",k8sServices:"all-services",k8sNamespace:"",k8sNamespaceSelection:"",k8sServiceDeployment:"",k8sServiceDeploymentSelection:"",k8sIngressDeployment:"",k8sIngressDeploymentSelection:"",k8sIngressType:"",k8sIngressBrand:"kong-ingress",k8sIngressSelection:""}),z=U(()=>{const d=Object.assign({},q.value),t=n.value.k8sNamespaceSelection;if(!t)return"";d.metadata.name=t,d.metadata.namespace=t,d.metadata.annotations["kuma.io/mesh"]=n.value.meshName;const o=`" | kubectl apply -f - && kubectl delete pod --all -n ${t}`;return se(d,o)}),A=U(()=>{const{k8sNamespaceSelection:d,meshName:t}=n.value;return t.length===0?!0:D.value===1?!d:!1});R(()=>n.value.k8sNamespaceSelection,function(d){n.value.k8sNamespaceSelection=ie(d)});const T=pe.get("step");D.value=T!==null?parseInt(T):0;function W(d){D.value=d}function $(){I.value=!0}async function G(){const t=n.value.meshName,o=n.value.k8sNamespaceSelection;if(V.value=!1,w.value=!1,!(!t||!o))try{const f=await P.getDataplaneFromMesh({mesh:t,name:o});f&&f.name.length>0?M.value=!0:w.value=!0}catch(f){w.value=!0,console.error(f)}finally{V.value=!0}}function O(){x.dispatch("updateSelectedMesh",n.value.meshName),j.push({name:"data-planes-list-view",params:{mesh:n.value.meshName}})}return(d,t)=>(u(),S(de,null,{default:s(()=>[i(ce,{title:r(k)("wizard-kubernetes.routes.item.title")},null,8,["title"]),a(),i(re,null,{default:s(()=>[e("div",he,[e("div",ve,[i(ne,{steps:K,"sidebar-content":F,"footer-enabled":I.value===!1,"next-disabled":A.value,onGoToStep:W},{general:s(()=>[_e,a(),e("p",null,`
                Welcome to the wizard to create a new Dataplane resource in `+m(r(k)("common.product.name"))+`.
                We will be providing you with a few steps that will get you started.
              `,1),a(),e("p",null,`
                As you know, the `+m(r(k)("common.product.name"))+` GUI is read-only.
              `,1),a(),ke,a(),ye,a(),ge,a(),i(r(v),{class:"my-6","has-shadow":""},{body:s(()=>[i(_,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:s(()=>[e("div",null,[c(e("select",{id:"dp-mesh","onUpdate:modelValue":t[0]||(t[0]=o=>n.value.meshName=o),class:"k-input w-100"},[fe,a(),(u(!0),h(H,null,Z(r(x).getters.getMeshList.items,o=>(u(),h("option",{key:o.name,value:o.name},m(o.name),9,be))),128))],512),[[Q,n.value.meshName]])]),a(),e("div",null,[we,a(),i(r(B),{to:{name:"create-mesh"},appearance:"outline"},{default:s(()=>[a(`
                        Create a new Mesh
                      `)]),_:1})])]),_:1})]),_:1})]),"scope-settings":s(()=>[Se,a(),De,a(),i(r(v),{class:"my-6","has-shadow":""},{body:s(()=>[i(_,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:s(()=>[e("label",Ie,[c(e("input",{id:"service-dataplane","onUpdate:modelValue":t[1]||(t[1]=o=>n.value.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[y,n.value.k8sDataplaneType]]),a(),Ne]),a(),e("label",xe,[c(e("input",{id:"ingress-dataplane","onUpdate:modelValue":t[2]||(t[2]=o=>n.value.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-ingress",disabled:""},null,512),[[y,n.value.k8sDataplaneType]]),a(),Me])]),_:1})]),_:1}),a(),n.value.k8sDataplaneType==="dataplane-type-service"?(u(),h("div",Ve,[Te,a(),i(r(v),{class:"my-6","has-shadow":""},{body:s(()=>[i(_,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:s(()=>[e("label",Ue,[c(e("input",{id:"k8s-services-all","onUpdate:modelValue":t[3]||(t[3]=o=>n.value.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"all-services",checked:""},null,512),[[y,n.value.k8sServices]]),a(),Be]),a(),e("label",Ce,[c(e("input",{id:"k8s-services-individual","onUpdate:modelValue":t[4]||(t[4]=o=>n.value.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"individual-services",disabled:""},null,512),[[y,n.value.k8sServices]]),a(),Ee])]),_:1})]),_:1}),a(),n.value.k8sServices==="individual-services"?(u(),S(r(v),{key:0,class:"my-6","has-shadow":""},{body:s(()=>[i(_,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:s(()=>[c(e("input",{id:"k8s-service-deployment-new","onUpdate:modelValue":t[5]||(t[5]=o=>n.value.k8sServiceDeploymentSelection=o),type:"text",class:"k-input w-100",placeholder:"your-new-deployment",required:""},null,512),[[N,n.value.k8sServiceDeploymentSelection]])]),_:1})]),_:1})):g("",!0),a(),i(r(v),{class:"my-6","has-shadow":""},{body:s(()=>[i(_,{title:"Namespace","for-attr":"k8s-namespace-selection"},{default:s(()=>[c(e("input",{id:"k8s-namespace-new","onUpdate:modelValue":t[6]||(t[6]=o=>n.value.k8sNamespaceSelection=o),type:"text",class:"k-input w-100",placeholder:"your-namespace",required:""},null,512),[[N,n.value.k8sNamespaceSelection]])]),_:1})]),_:1})])):g("",!0),a(),n.value.k8sDataplaneType==="dataplane-type-ingress"?(u(),h("div",Pe,[e("p",null,m(r(k)("common.product.name"))+` natively supports the Kong Ingress. Do you want to deploy
                  Kong or another Ingress?
                `,1),a(),i(r(v),{class:"my-6","has-shadow":""},{body:s(()=>[i(_,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:s(()=>[e("label",Ke,[c(e("input",{id:"k8s-ingress-kong","onUpdate:modelValue":t[7]||(t[7]=o=>n.value.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"kong-ingress",checked:""},null,512),[[y,n.value.k8sIngressBrand]]),a(),Fe]),a(),e("label",je,[c(e("input",{id:"k8s-ingress-other","onUpdate:modelValue":t[8]||(t[8]=o=>n.value.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"other-ingress"},null,512),[[y,n.value.k8sIngressBrand]]),a(),qe])]),_:1})]),_:1}),a(),i(r(v),{class:"my-6","has-shadow":""},{body:s(()=>[i(_,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:s(()=>[c(e("input",{id:"k8s-ingress-deployment-new","onUpdate:modelValue":t[9]||(t[9]=o=>n.value.k8sIngressDeployment=o),type:"text",class:"k-input w-100",placeholder:"your-deployment",required:""},null,512),[[N,n.value.k8sIngressDeployment]])]),_:1})]),_:1}),a(),n.value.k8sIngressBrand==="other-ingress"?(u(),S(r(C),{key:0,appearance:"info"},{alertMessage:s(()=>[ze]),_:1})):g("",!0)])):g("",!0)]),complete:s(()=>[n.value.meshName?(u(),h("div",Ae,[I.value===!1?(u(),h("div",We,[$e,a(),Ge,a(),Oe,a(),i(E,{id:"code-block-kubernetes-command",class:"mt-3",language:"bash",code:z.value},null,8,["code"])])):g("",!0),a(),i(ee,{"loader-function":G,"has-error":w.value,"can-complete":M.value,onHideSiblings:$},{"loading-title":s(()=>[Le]),"loading-content":s(()=>[Ye]),"complete-title":s(()=>[Re]),"complete-content":s(()=>[e("p",null,[a(`
                      Your Dataplane
                      `),n.value.k8sNamespaceSelection?(u(),h("strong",Qe,m(n.value.k8sNamespaceSelection),1)):g("",!0),a(`
                      was found!
                    `)]),a(),Ze,a(),e("p",null,[i(r(B),{appearance:"primary",onClick:O},{default:s(()=>[a(`
                        View Your Dataplane
                      `)]),_:1})])]),"error-title":s(()=>[He]),"error-content":s(()=>[Je]),_:1},8,["has-error","can-complete"])])):(u(),S(r(C),{key:1,appearance:"danger"},{alertMessage:s(()=>[Xe]),_:1}))]),dataplane:s(()=>[ea,a(),e("p",null,`
                In `+m(r(k)("common.product.name"))+`, a Dataplane resource represents a data plane proxy running
                alongside one of your services. Data plane proxies can be added in any Mesh
                that you may have created, and in Kubernetes, they will be auto-injected
                by `+m(r(k)("common.product.name"))+`.
              `,1)]),example:s(()=>[aa,a(),na,a(),i(E,{id:"onboarding-dpp-kubernetes-example",class:"sample-code-block",code:sa,language:"yaml"})]),switch:s(()=>[i(ae)]),_:1},8,["footer-enabled","next-disabled"])])])]),_:1})]),_:1}))}});const _a=ue(ta,[["__scopeId","data-v-6e5d7460"]]);export{_a as default};
