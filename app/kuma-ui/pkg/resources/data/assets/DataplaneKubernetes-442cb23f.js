import{m as x}from"./vuex.esm-bundler-df5bd11e.js";import{a as M,P as B,D as K}from"./kongponents.es-3df60cd6.js";import{y as V,k as F}from"./store-d743b198.js";import{_ as P}from"./CodeBlock.vue_vue_type_style_index_0_lang-e4b856ed.js";import{f as U}from"./formatForCLI-931cd5c6.js";import{F as T,S as A,E as q}from"./EntityScanner-9b947b61.js";import{E as z}from"./EnvironmentSwitcher-03d5770b.js";import{P as O}from"./constants-31fdaf55.js";import{l as p,h as u,g as e,e as r,w as s,o as d,f as n,t as k,S as c,a2 as W,F as j,v as L,a4 as v,a as b,a1 as S,b as y,p as G,m as Y}from"./runtime-dom.esm-bundler-91b41870.js";import{_ as R}from"./_plugin-vue_export-helper-c27b6911.js";import"./_commonjsHelpers-87174ba5.js";import"./toYaml-4e00099e.js";const H={apiVersion:"v1",kind:"Namespace",metadata:{name:null,namespace:null,annotations:{"kuma.io/sidecar-injection":"enabled","kuma.io/mesh":null}}};const X=`apiVersion: 'kuma.io/v1alpha1'
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
      kuma.io/service: echo`,J={name:"DataplaneWizardKubernetes",EXAMPLE_CODE:X,components:{CodeBlock:P,FormFragment:T,StepSkeleton:A,EnvironmentSwitcher:z,EntityScanner:q,KAlert:M,KButton:B,KCard:K},data(){return{productName:O,schema:H,steps:[{label:"General",slug:"general"},{label:"Scope Settings",slug:"scope-settings"},{label:"Install",slug:"complete"}],tabs:[{hash:"#kubernetes",title:"Kubernetes"}],sidebarContent:[{name:"dataplane"},{name:"example"},{name:"switch"}],startScanner:!1,scanFound:!1,hideScannerSiblings:!1,scanError:!1,isComplete:!1,validate:{meshName:"",k8sDataplaneType:"dataplane-type-service",k8sServices:"all-services",k8sNamespace:"",k8sNamespaceSelection:"",k8sServiceDeployment:"",k8sServiceDeploymentSelection:"",k8sIngressDeployment:"",k8sIngressDeploymentSelection:"",k8sIngressType:"",k8sIngressBrand:"kong-ingress",k8sIngressSelection:""}}},computed:{...x({title:"config/getTagline",version:"config/getVersion",environment:"config/getEnvironment",meshes:"getMeshList"}),codeOutput(){const i=Object.assign({},this.schema),a=this.validate.k8sNamespaceSelection;if(!a)return;i.metadata.name=a,i.metadata.namespace=a,i.metadata.annotations["kuma.io/mesh"]=this.validate.meshName;const g=`" | kubectl apply -f - && kubectl delete pod --all -n ${a}`;return U(i,g)},nextDisabled(){const{k8sNamespaceSelection:i,meshName:a}=this.validate;return a.length?this.$route.query.step==="1"?!i:!1:!0}},watch:{"validate.k8sNamespaceSelection"(i){this.validate.k8sNamespaceSelection=V(i)},$route(){this.$route.query.step===1&&(this.validate.k8sNamespaceSelection?this.nextDisabled=!1:this.nextDisabled=!0)}},methods:{hideSiblings(){this.hideScannerSiblings=!0},scanForEntity(){const a=this.validate.meshName,g=this.validate.k8sNamespaceSelection;this.scanComplete=!1,this.scanError=!1,!(!a||!g)&&F.getDataplaneFromMesh({mesh:a,name:g}).then(_=>{_&&_.name.length>0?(this.isRunning=!0,this.scanFound=!0):this.scanError=!0}).catch(_=>{this.scanError=!0,console.error(_)}).finally(()=>{this.scanComplete=!0})},compeleteDataPlaneSetup(){this.$store.dispatch("updateSelectedMesh",this.validate.meshName),this.$router.push({name:"data-plane-list-view",params:{mesh:this.validate.meshName}})}}},l=i=>(G("data-v-a9f9011e"),i=i(),Y(),i),Q={class:"wizard"},Z={class:"wizard__content"},$=l(()=>e("h3",null,`
            Create Kubernetes Dataplane
          `,-1)),ee=l(()=>e("h3",null,`
            To get started, please select on what Mesh you would like to add the Dataplane:
          `,-1)),ne=l(()=>e("p",null,`
            If you've got an existing Mesh that you would like to associate with your
            Dataplane, you can select it below, or create a new one using our Mesh Wizard.
          `,-1)),te=l(()=>e("small",null,"Would you like to see instructions for Universal? Use sidebar to change wizard!",-1)),ae=l(()=>e("option",{disabled:"",value:""},`
                      Select an existing Mesh…
                    `,-1)),se=["value"],le=l(()=>e("label",{class:"k-input-label mr-4"},`
                    or
                  `,-1)),oe=l(()=>e("h3",null,`
            Setup Dataplane Mode
          `,-1)),ie=l(()=>e("p",null,`
            You can create a data plane for a service or a data plane for a Gateway.
          `,-1)),re={for:"service-dataplane"},de=l(()=>e("span",null,`
                    Service Dataplane
                  `,-1)),ce={for:"ingress-dataplane"},pe=l(()=>e("span",null,`
                    Ingress Dataplane
                  `,-1)),ue={key:0},me=l(()=>e("p",null,`
              Should the data plane be added for an entire Namespace and all of its services,
              or for specific individual services in any namespace?
            `,-1)),he={for:"k8s-services-all"},ke=l(()=>e("span",null,`
                      All Services in Namespace
                    `,-1)),_e={for:"k8s-services-individual"},ve=l(()=>e("span",null,`
                      Individual Services
                    `,-1)),ye={key:1},ge={for:"k8s-ingress-kong"},fe=l(()=>e("span",null,`
                      Kong Ingress
                    `,-1)),be={for:"k8s-ingress-other"},Se=l(()=>e("span",null,`
                      Other Ingress
                    `,-1)),we=l(()=>e("p",null,`
                  Please go ahead and deploy the Ingress first, then restart this wizard and select “Existing Ingress”.
                `,-1)),De={key:0},Ne={key:0},Ie=l(()=>e("h3",null,`
                Auto-Inject DPP
              `,-1)),Ee=l(()=>e("p",null,`
                You can now execute the following commands to automatically inject the sidecar proxy in every Pod, and by doing so creating the Dataplane.
              `,-1)),Ce=l(()=>e("h4",null,"Kubernetes",-1)),xe=l(()=>e("h3",null,"Searching…",-1)),Me=l(()=>e("p",null,"We are looking for your dataplane.",-1)),Be=l(()=>e("h3",null,"Done!",-1)),Ke={key:0},Ve=l(()=>e("p",null,`
                  Proceed to the next step where we will show you
                  your new Dataplane.
                `,-1)),Fe=l(()=>e("h3",null,"Mesh not found",-1)),Pe=l(()=>e("p",null,"We were unable to find your mesh.",-1)),Ue=l(()=>e("p",null,`
                Please return to the first step and make sure to select an
                existing Mesh, or create a new one.
              `,-1)),Te=l(()=>e("h3",null,"Dataplane",-1)),Ae=l(()=>e("h3",null,"Example",-1)),qe=l(()=>e("p",null,`
            Below is an example of a Dataplane resource output:
          `,-1));function ze(i,a,g,_,t,f){const w=p("KButton"),m=p("FormFragment"),h=p("KCard"),D=p("KAlert"),N=p("CodeBlock"),I=p("EntityScanner"),E=p("EnvironmentSwitcher"),C=p("StepSkeleton");return d(),u("div",Q,[e("div",Z,[r(C,{steps:t.steps,"sidebar-content":t.sidebarContent,"footer-enabled":t.hideScannerSiblings===!1,"next-disabled":f.nextDisabled},{general:s(()=>[$,n(),e("p",null,`
            Welcome to the wizard to create a new Dataplane resource in `+k(i.title)+`.
            We will be providing you with a few steps that will get you started.
          `,1),n(),e("p",null,`
            As you know, the `+k(t.productName)+` GUI is read-only.
          `,1),n(),ee,n(),ne,n(),te,n(),r(h,{class:"my-6","has-shadow":""},{body:s(()=>[r(m,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:s(()=>[e("div",null,[c(e("select",{id:"dp-mesh","onUpdate:modelValue":a[0]||(a[0]=o=>t.validate.meshName=o),class:"k-input w-100"},[ae,n(),(d(!0),u(j,null,L(i.meshes.items,o=>(d(),u("option",{key:o.name,value:o.name},k(o.name),9,se))),128))],512),[[W,t.validate.meshName]])]),n(),e("div",null,[le,n(),r(w,{to:{name:"create-mesh"},appearance:"outline"},{default:s(()=>[n(`
                    Create a new Mesh
                  `)]),_:1})])]),_:1})]),_:1})]),"scope-settings":s(()=>[oe,n(),ie,n(),r(h,{class:"my-6","has-shadow":""},{body:s(()=>[r(m,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:s(()=>[e("label",re,[c(e("input",{id:"service-dataplane","onUpdate:modelValue":a[1]||(a[1]=o=>t.validate.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[v,t.validate.k8sDataplaneType]]),n(),de]),n(),e("label",ce,[c(e("input",{id:"ingress-dataplane","onUpdate:modelValue":a[2]||(a[2]=o=>t.validate.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-ingress",disabled:""},null,512),[[v,t.validate.k8sDataplaneType]]),n(),pe])]),_:1})]),_:1}),n(),t.validate.k8sDataplaneType==="dataplane-type-service"?(d(),u("div",ue,[me,n(),r(h,{class:"my-6","has-shadow":""},{body:s(()=>[r(m,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:s(()=>[e("label",he,[c(e("input",{id:"k8s-services-all","onUpdate:modelValue":a[3]||(a[3]=o=>t.validate.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"all-services",checked:""},null,512),[[v,t.validate.k8sServices]]),n(),ke]),n(),e("label",_e,[c(e("input",{id:"k8s-services-individual","onUpdate:modelValue":a[4]||(a[4]=o=>t.validate.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"individual-services",disabled:""},null,512),[[v,t.validate.k8sServices]]),n(),ve])]),_:1})]),_:1}),n(),t.validate.k8sServices==="individual-services"?(d(),b(h,{key:0,class:"my-6","has-shadow":""},{body:s(()=>[r(m,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:s(()=>[c(e("input",{id:"k8s-service-deployment-new","onUpdate:modelValue":a[5]||(a[5]=o=>t.validate.k8sServiceDeploymentSelection=o),type:"text",class:"k-input w-100",placeholder:"your-new-deployment",required:""},null,512),[[S,t.validate.k8sServiceDeploymentSelection]])]),_:1})]),_:1})):y("",!0),n(),r(h,{class:"my-6","has-shadow":""},{body:s(()=>[r(m,{title:"Namespace","for-attr":"k8s-namespace-selection"},{default:s(()=>[c(e("input",{id:"k8s-namespace-new","onUpdate:modelValue":a[6]||(a[6]=o=>t.validate.k8sNamespaceSelection=o),type:"text",class:"k-input w-100",placeholder:"your-namespace",required:""},null,512),[[S,t.validate.k8sNamespaceSelection]])]),_:1})]),_:1})])):y("",!0),n(),t.validate.k8sDataplaneType==="dataplane-type-ingress"?(d(),u("div",ye,[e("p",null,k(i.title)+` natively supports the Kong Ingress. Do you want to deploy
              Kong or another Ingress?
            `,1),n(),r(h,{class:"my-6","has-shadow":""},{body:s(()=>[r(m,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:s(()=>[e("label",ge,[c(e("input",{id:"k8s-ingress-kong","onUpdate:modelValue":a[7]||(a[7]=o=>t.validate.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"kong-ingress",checked:""},null,512),[[v,t.validate.k8sIngressBrand]]),n(),fe]),n(),e("label",be,[c(e("input",{id:"k8s-ingress-other","onUpdate:modelValue":a[8]||(a[8]=o=>t.validate.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"other-ingress"},null,512),[[v,t.validate.k8sIngressBrand]]),n(),Se])]),_:1})]),_:1}),n(),r(h,{class:"my-6","has-shadow":""},{body:s(()=>[r(m,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:s(()=>[c(e("input",{id:"k8s-ingress-deployment-new","onUpdate:modelValue":a[9]||(a[9]=o=>t.validate.k8sIngressDeployment=o),type:"text",class:"k-input w-100",placeholder:"your-deployment",required:""},null,512),[[S,t.validate.k8sIngressDeployment]])]),_:1})]),_:1}),n(),t.validate.k8sIngressBrand==="other-ingress"?(d(),b(D,{key:0,appearance:"info"},{alertMessage:s(()=>[we]),_:1})):y("",!0)])):y("",!0)]),complete:s(()=>[t.validate.meshName?(d(),u("div",De,[t.hideScannerSiblings===!1?(d(),u("div",Ne,[Ie,n(),Ee,n(),Ce,n(),r(N,{id:"code-block-kubernetes-command",class:"mt-3",language:"bash",code:f.codeOutput},null,8,["code"])])):y("",!0),n(),r(I,{"loader-function":f.scanForEntity,"should-start":!0,"has-error":t.scanError,"can-complete":t.scanFound,onHideSiblings:f.hideSiblings},{"loading-title":s(()=>[xe]),"loading-content":s(()=>[Me]),"complete-title":s(()=>[Be]),"complete-content":s(()=>[e("p",null,[n(`
                  Your Dataplane
                  `),t.validate.k8sNamespaceSelection?(d(),u("strong",Ke,k(t.validate.k8sNamespaceSelection),1)):y("",!0),n(`
                  was found!
                `)]),n(),Ve,n(),e("p",null,[r(w,{appearance:"primary",onClick:f.compeleteDataPlaneSetup},{default:s(()=>[n(`
                    View Your Dataplane
                  `)]),_:1},8,["onClick"])])]),"error-title":s(()=>[Fe]),"error-content":s(()=>[Pe]),_:1},8,["loader-function","has-error","can-complete","onHideSiblings"])])):(d(),b(D,{key:1,appearance:"danger"},{alertMessage:s(()=>[Ue]),_:1}))]),dataplane:s(()=>[Te,n(),e("p",null,`
            In `+k(i.title)+`, a Dataplane resource represents a data plane proxy running
            alongside one of your services. Data plane proxies can be added in any Mesh
            that you may have created, and in Kubernetes, they will be auto-injected
            by `+k(i.title)+`.
          `,1)]),example:s(()=>[Ae,n(),qe,n(),r(N,{id:"onboarding-dpp-kubernetes-example",class:"sample-code-block",code:i.$options.EXAMPLE_CODE,language:"yaml"},null,8,["code"])]),switch:s(()=>[r(E)]),_:1},8,["steps","sidebar-content","footer-enabled","next-disabled"])])])}const $e=R(J,[["render",ze],["__scopeId","data-v-a9f9011e"]]);export{$e as default};
