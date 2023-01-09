import{H as x,z as M,A as B,O as K,cl as V,co as F,cQ as U,k as T,i as p,o as d,j as u,l as e,a as r,w as a,b as n,t as k,m as c,cF as A,F as P,n as q,cR as v,c as f,v as S,B as y,E as z,G as O}from"./index.e014f0d3.js";import{_ as j}from"./CodeBlock.vue_vue_type_style_index_0_lang.e945f8ce.js";import{f as W}from"./formatForCLI.199be697.js";import{F as G,S as L,E as R}from"./EntityScanner.65564897.js";import{E as Y}from"./EnvironmentSwitcher.331977d3.js";import"./_commonjsHelpers.f037b798.js";import"./index.58caa11d.js";const H={apiVersion:"v1",kind:"Namespace",metadata:{name:null,namespace:null,annotations:{"kuma.io/sidecar-injection":"enabled","kuma.io/mesh":null}}};const X=`apiVersion: 'kuma.io/v1alpha1'
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
      kuma.io/service: echo`,Q={name:"DataplaneWizardKubernetes",EXAMPLE_CODE:X,components:{CodeBlock:j,FormFragment:G,StepSkeleton:L,EnvironmentSwitcher:Y,EntityScanner:R,KAlert:M,KButton:B,KCard:K},data(){return{productName:V,schema:H,steps:[{label:"General",slug:"general"},{label:"Scope Settings",slug:"scope-settings"},{label:"Install",slug:"complete"}],tabs:[{hash:"#kubernetes",title:"Kubernetes"}],sidebarContent:[{name:"dataplane"},{name:"example"},{name:"switch"}],startScanner:!1,scanFound:!1,hideScannerSiblings:!1,scanError:!1,isComplete:!1,validate:{meshName:"",k8sDataplaneType:"dataplane-type-service",k8sServices:"all-services",k8sNamespace:"",k8sNamespaceSelection:"",k8sServiceDeployment:"",k8sServiceDeploymentSelection:"",k8sIngressDeployment:"",k8sIngressDeploymentSelection:"",k8sIngressType:"",k8sIngressBrand:"kong-ingress",k8sIngressSelection:""}}},computed:{...F({title:"config/getTagline",version:"config/getVersion",environment:"config/getEnvironment",meshes:"getMeshList"}),codeOutput(){const i=Object.assign({},this.schema),s=this.validate.k8sNamespaceSelection;if(!s)return;i.metadata.name=s,i.metadata.namespace=s,i.metadata.annotations["kuma.io/mesh"]=this.validate.meshName;const g=`" | kubectl apply -f - && kubectl delete pod --all -n ${s}`;return W(i,g)},nextDisabled(){const{k8sNamespaceSelection:i,meshName:s}=this.validate;return s.length?this.$route.query.step==="1"?!i:!1:!0}},watch:{"validate.k8sNamespaceSelection"(i){this.validate.k8sNamespaceSelection=U(i)},$route(){this.$route.query.step===1&&(this.validate.k8sNamespaceSelection?this.nextDisabled=!1:this.nextDisabled=!0)}},methods:{hideSiblings(){this.hideScannerSiblings=!0},scanForEntity(){const s=this.validate.meshName,g=this.validate.k8sNamespaceSelection;this.scanComplete=!1,this.scanError=!1,!(!s||!g)&&T.getDataplaneFromMesh({mesh:s,name:g}).then(_=>{_&&_.name.length>0?(this.isRunning=!0,this.scanFound=!0):this.scanError=!0}).catch(_=>{this.scanError=!0,console.error(_)}).finally(()=>{this.scanComplete=!0})},compeleteDataPlaneSetup(){this.$store.dispatch("updateSelectedMesh",this.validate.meshName),this.$router.push({name:"data-plane-list-view",params:{mesh:this.validate.meshName}})}}},l=i=>(z("data-v-e050c731"),i=i(),O(),i),J={class:"wizard"},Z={class:"wizard__content"},$=l(()=>e("h3",null,`
            Create Kubernetes Dataplane
          `,-1)),ee=l(()=>e("h3",null,`
            To get started, please select on what Mesh you would like to add the Dataplane:
          `,-1)),ne=l(()=>e("p",null,`
            If you've got an existing Mesh that you would like to associate with your
            Dataplane, you can select it below, or create a new one using our Mesh Wizard.
          `,-1)),te=l(()=>e("small",null,"Would you like to see instructions for Universal? Use sidebar to change wizard!",-1)),se=l(()=>e("option",{disabled:"",value:""},`
                      Select an existing Mesh\u2026
                    `,-1)),ae=["value"],le=l(()=>e("label",{class:"k-input-label mr-4"},`
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
                    `,-1)),ye={key:1},ge={for:"k8s-ingress-kong"},be=l(()=>e("span",null,`
                      Kong Ingress
                    `,-1)),fe={for:"k8s-ingress-other"},Se=l(()=>e("span",null,`
                      Other Ingress
                    `,-1)),we=l(()=>e("p",null,`
                  Please go ahead and deploy the Ingress first, then restart this
                  wizard and select "Existing Ingress".
                `,-1)),De={key:0},Ne={key:0},Ie=l(()=>e("h3",null,`
                Auto-Inject DPP
              `,-1)),Ee=l(()=>e("p",null,`
                You can now execute the following commands to automatically inject
                the sidecar proxy in every Pod, and by doing so creating the Dataplane.
              `,-1)),Ce=l(()=>e("h4",null,"Kubernetes",-1)),xe=l(()=>e("h3",null,"Searching\u2026",-1)),Me=l(()=>e("p",null,"We are looking for your dataplane.",-1)),Be=l(()=>e("h3",null,"Done!",-1)),Ke={key:0},Ve=l(()=>e("p",null,`
                  Proceed to the next step where we will show you
                  your new Dataplane.
                `,-1)),Fe=l(()=>e("h3",null,"Mesh not found",-1)),Ue=l(()=>e("p",null,"We were unable to find your mesh.",-1)),Te=l(()=>e("p",null,`
                Please return to the first step and make sure to select an
                existing Mesh, or create a new one.
              `,-1)),Ae=l(()=>e("h3",null,"Dataplane",-1)),Pe=l(()=>e("h3",null,"Example",-1)),qe=l(()=>e("p",null,`
            Below is an example of a Dataplane resource output:
          `,-1));function ze(i,s,g,_,t,b){const w=p("KButton"),m=p("FormFragment"),h=p("KCard"),D=p("KAlert"),N=p("CodeBlock"),I=p("EntityScanner"),E=p("EnvironmentSwitcher"),C=p("StepSkeleton");return d(),u("div",J,[e("div",Z,[r(C,{steps:t.steps,"sidebar-content":t.sidebarContent,"footer-enabled":t.hideScannerSiblings===!1,"next-disabled":b.nextDisabled},{general:a(()=>[$,n(),e("p",null,`
            Welcome to the wizard to create a new Dataplane resource in `+k(i.title)+`.
            We will be providing you with a few steps that will get you started.
          `,1),n(),e("p",null,`
            As you know, the `+k(t.productName)+` GUI is read-only.
          `,1),n(),ee,n(),ne,n(),te,n(),r(h,{class:"my-6","has-shadow":""},{body:a(()=>[r(m,{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""},{default:a(()=>[e("div",null,[c(e("select",{id:"dp-mesh","onUpdate:modelValue":s[0]||(s[0]=o=>t.validate.meshName=o),class:"k-input w-100"},[se,n(),(d(!0),u(P,null,q(i.meshes.items,o=>(d(),u("option",{key:o.name,value:o.name},k(o.name),9,ae))),128))],512),[[A,t.validate.meshName]])]),n(),e("div",null,[le,n(),r(w,{to:{name:"create-mesh"},appearance:"outline"},{default:a(()=>[n(`
                    Create a new Mesh
                  `)]),_:1})])]),_:1})]),_:1})]),"scope-settings":a(()=>[oe,n(),ie,n(),r(h,{class:"my-6","has-shadow":""},{body:a(()=>[r(m,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:a(()=>[e("label",re,[c(e("input",{id:"service-dataplane","onUpdate:modelValue":s[1]||(s[1]=o=>t.validate.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},null,512),[[v,t.validate.k8sDataplaneType]]),n(),de]),n(),e("label",ce,[c(e("input",{id:"ingress-dataplane","onUpdate:modelValue":s[2]||(s[2]=o=>t.validate.k8sDataplaneType=o),class:"k-input",type:"radio",name:"dataplane-type",value:"dataplane-type-ingress",disabled:""},null,512),[[v,t.validate.k8sDataplaneType]]),n(),pe])]),_:1})]),_:1}),n(),t.validate.k8sDataplaneType==="dataplane-type-service"?(d(),u("div",ue,[me,n(),r(h,{class:"my-6","has-shadow":""},{body:a(()=>[r(m,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:a(()=>[e("label",he,[c(e("input",{id:"k8s-services-all","onUpdate:modelValue":s[3]||(s[3]=o=>t.validate.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"all-services",checked:""},null,512),[[v,t.validate.k8sServices]]),n(),ke]),n(),e("label",_e,[c(e("input",{id:"k8s-services-individual","onUpdate:modelValue":s[4]||(s[4]=o=>t.validate.k8sServices=o),class:"k-input",type:"radio",name:"k8s-services",value:"individual-services",disabled:""},null,512),[[v,t.validate.k8sServices]]),n(),ve])]),_:1})]),_:1}),n(),t.validate.k8sServices==="individual-services"?(d(),f(h,{key:0,class:"my-6","has-shadow":""},{body:a(()=>[r(m,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:a(()=>[c(e("input",{id:"k8s-service-deployment-new","onUpdate:modelValue":s[5]||(s[5]=o=>t.validate.k8sServiceDeploymentSelection=o),type:"text",class:"k-input w-100",placeholder:"your-new-deployment",required:""},null,512),[[S,t.validate.k8sServiceDeploymentSelection]])]),_:1})]),_:1})):y("",!0),n(),r(h,{class:"my-6","has-shadow":""},{body:a(()=>[r(m,{title:"Namespace","for-attr":"k8s-namespace-selection"},{default:a(()=>[c(e("input",{id:"k8s-namespace-new","onUpdate:modelValue":s[6]||(s[6]=o=>t.validate.k8sNamespaceSelection=o),type:"text",class:"k-input w-100",placeholder:"your-namespace",required:""},null,512),[[S,t.validate.k8sNamespaceSelection]])]),_:1})]),_:1})])):y("",!0),n(),t.validate.k8sDataplaneType==="dataplane-type-ingress"?(d(),u("div",ye,[e("p",null,k(i.title)+` natively supports the Kong Ingress. Do you want to deploy
              Kong or another Ingress?
            `,1),n(),r(h,{class:"my-6","has-shadow":""},{body:a(()=>[r(m,{"all-inline":"","equal-cols":"","hide-label-col":""},{default:a(()=>[e("label",ge,[c(e("input",{id:"k8s-ingress-kong","onUpdate:modelValue":s[7]||(s[7]=o=>t.validate.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"kong-ingress",checked:""},null,512),[[v,t.validate.k8sIngressBrand]]),n(),be]),n(),e("label",fe,[c(e("input",{id:"k8s-ingress-other","onUpdate:modelValue":s[8]||(s[8]=o=>t.validate.k8sIngressBrand=o),class:"k-input",type:"radio",name:"k8s-ingress-brand",value:"other-ingress"},null,512),[[v,t.validate.k8sIngressBrand]]),n(),Se])]),_:1})]),_:1}),n(),r(h,{class:"my-6","has-shadow":""},{body:a(()=>[r(m,{title:"Deployments","for-attr":"k8s-deployment-selection"},{default:a(()=>[c(e("input",{id:"k8s-ingress-deployment-new","onUpdate:modelValue":s[9]||(s[9]=o=>t.validate.k8sIngressDeployment=o),type:"text",class:"k-input w-100",placeholder:"your-deployment",required:""},null,512),[[S,t.validate.k8sIngressDeployment]])]),_:1})]),_:1}),n(),t.validate.k8sIngressBrand==="other-ingress"?(d(),f(D,{key:0,appearance:"info"},{alertMessage:a(()=>[we]),_:1})):y("",!0)])):y("",!0)]),complete:a(()=>[t.validate.meshName?(d(),u("div",De,[t.hideScannerSiblings===!1?(d(),u("div",Ne,[Ie,n(),Ee,n(),Ce,n(),r(N,{id:"code-block-kubernetes-command",class:"mt-3",language:"bash",code:b.codeOutput},null,8,["code"])])):y("",!0),n(),r(I,{"loader-function":b.scanForEntity,"should-start":!0,"has-error":t.scanError,"can-complete":t.scanFound,onHideSiblings:b.hideSiblings},{"loading-title":a(()=>[xe]),"loading-content":a(()=>[Me]),"complete-title":a(()=>[Be]),"complete-content":a(()=>[e("p",null,[n(`
                  Your Dataplane
                  `),t.validate.k8sNamespaceSelection?(d(),u("strong",Ke,k(t.validate.k8sNamespaceSelection),1)):y("",!0),n(`
                  was found!
                `)]),n(),Ve,n(),e("p",null,[r(w,{appearance:"primary",onClick:b.compeleteDataPlaneSetup},{default:a(()=>[n(`
                    View Your Dataplane
                  `)]),_:1},8,["onClick"])])]),"error-title":a(()=>[Fe]),"error-content":a(()=>[Ue]),_:1},8,["loader-function","has-error","can-complete","onHideSiblings"])])):(d(),f(D,{key:1,appearance:"danger"},{alertMessage:a(()=>[Te]),_:1}))]),dataplane:a(()=>[Ae,n(),e("p",null,`
            In `+k(i.title)+`, a Dataplane resource represents a data plane proxy running
            alongside one of your services. Data plane proxies can be added in any Mesh
            that you may have created, and in Kubernetes, they will be auto-injected
            by `+k(i.title)+`.
          `,1)]),example:a(()=>[Pe,n(),qe,n(),r(N,{id:"onboarding-dpp-kubernetes-example",code:i.$options.EXAMPLE_CODE,language:"yaml"},null,8,["code"])]),switch:a(()=>[r(E)]),_:1},8,["steps","sidebar-content","footer-enabled","next-disabled"])])])}const He=x(Q,[["render",ze],["__scopeId","data-v-e050c731"]]);export{He as default};
