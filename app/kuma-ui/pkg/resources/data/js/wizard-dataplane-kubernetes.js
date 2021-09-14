(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["wizard-dataplane-kubernetes"],{1373:function(e,t,a){"use strict";a("99af");var n=a("e80b"),s=a.n(n);t["a"]={methods:{formatForCLI:function(e){var t=arguments.length>1&&void 0!==arguments[1]?arguments[1]:'" | kumactl apply -f -',a='echo "',n=s()(e);return"".concat(a).concat(n).concat(t)}}}},"8b9a":function(e,t,a){"use strict";a("ee3d")},a527:function(e,t,a){"use strict";a.r(t);var n=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{staticClass:"wizard"},[a("div",{staticClass:"wizard__content"},[a("StepSkeleton",{attrs:{steps:e.steps,"sidebar-content":e.sidebarContent,"footer-enabled":!1===e.hideScannerSiblings,"next-disabled":e.nextDisabled},scopedSlots:e._u([{key:"general",fn:function(){return[a("h3",[e._v(" Create Kubernetes Dataplane ")]),a("p",[e._v(" Welcome to the wizard to create a new Dataplane resource in "+e._s(e.title)+". We will be providing you with a few steps that will get you started. ")]),a("p",[e._v(" As you know, the "+e._s(e.productName)+" GUI is read-only. ")]),a("h3",[e._v(" To get started, please select on what Mesh you would like to add the Dataplane: ")]),a("p",[e._v(" If you've got an existing Mesh that you would like to associate with your Dataplane, you can select it below, or create a new one using our Mesh Wizard. ")]),a("small",[e._v("Would you like to see instructions for Universal? Use sidebar to change wizard!")]),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""},scopedSlots:e._u([{key:"body",fn:function(){return[a("FormFragment",{attrs:{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""}},[a("div",[a("select",{directives:[{name:"model",rawName:"v-model",value:e.validate.meshName,expression:"validate.meshName"}],staticClass:"k-input w-100",attrs:{id:"dp-mesh"},on:{change:function(t){var a=Array.prototype.filter.call(t.target.options,(function(e){return e.selected})).map((function(e){var t="_value"in e?e._value:e.value;return t}));e.$set(e.validate,"meshName",t.target.multiple?a:a[0])}}},[a("option",{attrs:{disabled:"",value:""}},[e._v(" Select an existing Mesh… ")]),e._l(e.meshes.items,(function(t){return a("option",{key:t.name,domProps:{value:t.name}},[e._v(" "+e._s(t.name)+" ")])}))],2)]),a("div",[a("label",{staticClass:"k-input-label mr-4"},[e._v(" or ")]),a("KButton",{attrs:{to:{name:"create-mesh"},appearance:"secondary"}},[e._v(" Create a new Mesh ")])],1)])]},proxy:!0}])})]},proxy:!0},{key:"scope-settings",fn:function(){return[a("h3",[e._v(" Setup Dataplane Mode ")]),a("p",[e._v(" You can create a data plane for a service or a data plane for a Gateway. ")]),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""},scopedSlots:e._u([{key:"body",fn:function(){return[a("FormFragment",{attrs:{"all-inline":"","equal-cols":"","hide-label-col":""}},[a("label",{attrs:{for:"service-dataplane"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sDataplaneType,expression:"validate.k8sDataplaneType"}],staticClass:"k-input",attrs:{id:"service-dataplane",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},domProps:{checked:e._q(e.validate.k8sDataplaneType,"dataplane-type-service")},on:{change:function(t){return e.$set(e.validate,"k8sDataplaneType","dataplane-type-service")}}}),a("span",[e._v(" Service Dataplane ")])]),a("label",{attrs:{for:"ingress-dataplane"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sDataplaneType,expression:"validate.k8sDataplaneType"}],staticClass:"k-input",attrs:{id:"ingress-dataplane",type:"radio",name:"dataplane-type",value:"dataplane-type-ingress",disabled:""},domProps:{checked:e._q(e.validate.k8sDataplaneType,"dataplane-type-ingress")},on:{change:function(t){return e.$set(e.validate,"k8sDataplaneType","dataplane-type-ingress")}}}),a("span",[e._v(" Ingress Dataplane ")])])])]},proxy:!0}])}),"dataplane-type-service"===e.validate.k8sDataplaneType?a("div",[a("p",[e._v(" Should the data plane be added for an entire Namespace and all of its services, or for specific individual services in any namespace? ")]),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""},scopedSlots:e._u([{key:"body",fn:function(){return[a("FormFragment",{attrs:{"all-inline":"","equal-cols":"","hide-label-col":""}},[a("label",{attrs:{for:"k8s-services-all"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sServices,expression:"validate.k8sServices"}],staticClass:"k-input",attrs:{id:"k8s-services-all",type:"radio",name:"k8s-services",value:"all-services",checked:""},domProps:{checked:e._q(e.validate.k8sServices,"all-services")},on:{change:function(t){return e.$set(e.validate,"k8sServices","all-services")}}}),a("span",[e._v(" All Services in Namespace ")])]),a("label",{attrs:{for:"k8s-services-individual"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sServices,expression:"validate.k8sServices"}],staticClass:"k-input",attrs:{id:"k8s-services-individual",type:"radio",name:"k8s-services",value:"individual-services",disabled:""},domProps:{checked:e._q(e.validate.k8sServices,"individual-services")},on:{change:function(t){return e.$set(e.validate,"k8sServices","individual-services")}}}),a("span",[e._v(" Individual Services ")])])])]},proxy:!0}],null,!1,2127996134)}),"individual-services"===e.validate.k8sServices?a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""},scopedSlots:e._u([{key:"body",fn:function(){return[a("FormFragment",{attrs:{title:"Deployments","for-attr":"k8s-deployment-selection"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sServiceDeploymentSelection,expression:"validate.k8sServiceDeploymentSelection"}],staticClass:"k-input w-100",attrs:{id:"k8s-service-deployment-new",type:"text",placeholder:"your-new-deployment",required:""},domProps:{value:e.validate.k8sServiceDeploymentSelection},on:{input:function(t){t.target.composing||e.$set(e.validate,"k8sServiceDeploymentSelection",t.target.value)}}})])]},proxy:!0}],null,!1,1626108368)}):e._e(),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""},scopedSlots:e._u([{key:"body",fn:function(){return[a("FormFragment",{attrs:{title:"Namespace","for-attr":"k8s-namespace-selection"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sNamespaceSelection,expression:"validate.k8sNamespaceSelection"}],staticClass:"k-input w-100",attrs:{id:"k8s-namespace-new",type:"text",placeholder:"your-namespace",required:""},domProps:{value:e.validate.k8sNamespaceSelection},on:{input:function(t){t.target.composing||e.$set(e.validate,"k8sNamespaceSelection",t.target.value)}}})])]},proxy:!0}],null,!1,771225282)})],1):e._e(),"dataplane-type-ingress"===e.validate.k8sDataplaneType?a("div",[a("p",[e._v(" "+e._s(e.title)+" natively supports the Kong Ingress. Do you want to deploy Kong or another Ingress? ")]),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""},scopedSlots:e._u([{key:"body",fn:function(){return[a("FormFragment",{attrs:{"all-inline":"","equal-cols":"","hide-label-col":""}},[a("label",{attrs:{for:"k8s-ingress-kong"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sIngressBrand,expression:"validate.k8sIngressBrand"}],staticClass:"k-input",attrs:{id:"k8s-ingress-kong",type:"radio",name:"k8s-ingress-brand",value:"kong-ingress",checked:""},domProps:{checked:e._q(e.validate.k8sIngressBrand,"kong-ingress")},on:{change:function(t){return e.$set(e.validate,"k8sIngressBrand","kong-ingress")}}}),a("span",[e._v(" Kong Ingress ")])]),a("label",{attrs:{for:"k8s-ingress-other"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sIngressBrand,expression:"validate.k8sIngressBrand"}],staticClass:"k-input",attrs:{id:"k8s-ingress-other",type:"radio",name:"k8s-ingress-brand",value:"other-ingress"},domProps:{checked:e._q(e.validate.k8sIngressBrand,"other-ingress")},on:{change:function(t){return e.$set(e.validate,"k8sIngressBrand","other-ingress")}}}),a("span",[e._v(" Other Ingress ")])])])]},proxy:!0}],null,!1,1060751940)}),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""},scopedSlots:e._u([{key:"body",fn:function(){return[a("FormFragment",{attrs:{title:"Deployments","for-attr":"k8s-deployment-selection"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sIngressDeployment,expression:"validate.k8sIngressDeployment"}],staticClass:"k-input w-100",attrs:{id:"k8s-ingress-deployment-new",type:"text",placeholder:"your-deployment",required:""},domProps:{value:e.validate.k8sIngressDeployment},on:{input:function(t){t.target.composing||e.$set(e.validate,"k8sIngressDeployment",t.target.value)}}})])]},proxy:!0}],null,!1,1817964619)}),"other-ingress"===e.validate.k8sIngressBrand?a("KAlert",{attrs:{appearance:"info"},scopedSlots:e._u([{key:"alertMessage",fn:function(){return[a("p",[e._v(' Please go ahead and deploy the Ingress first, then restart this wizard and select "Existing Ingress". ')])]},proxy:!0}],null,!1,1402213972)}):e._e()],1):e._e()]},proxy:!0},{key:"complete",fn:function(){return[e.validate.meshName?a("div",[!1===e.hideScannerSiblings?a("div",[a("h3",[e._v(" Auto-Inject DPP ")]),a("p",[e._v(" You can now execute the following commands to automatically inject the sidecar proxy in every Pod, and by doing so creating the Dataplane. ")]),a("Tabs",{attrs:{loaders:!1,tabs:e.tabs,"has-border":!0,"initial-tab-override":"kubernetes"},scopedSlots:e._u([{key:"kubernetes",fn:function(){return[a("CodeView",{attrs:{title:"Kubernetes","copy-button-text":"Copy Command to Clipboard",lang:"bash",content:e.codeOutput}})]},proxy:!0}],null,!1,525752398)})],1):e._e(),a("Scanner",{attrs:{"loader-function":e.scanForEntity,"should-start":!0,"has-error":e.scanError,"can-complete":e.scanFound},on:{hideSiblings:e.hideSiblings},scopedSlots:e._u([{key:"loading-title",fn:function(){return[a("h3",[e._v("Searching…")])]},proxy:!0},{key:"loading-content",fn:function(){return[a("p",[e._v("We are looking for your dataplane.")])]},proxy:!0},{key:"complete-title",fn:function(){return[a("h3",[e._v("Done!")])]},proxy:!0},{key:"complete-content",fn:function(){return[a("p",[e._v(" Your Dataplane "),e.validate.k8sNamespaceSelection?a("strong",[e._v(" "+e._s(e.validate.k8sNamespaceSelection)+" ")]):e._e(),e._v(" was found! ")]),a("p",[e._v(" Proceed to the next step where we will show you your new Dataplane. ")]),a("p",[a("KButton",{attrs:{appearance:"primary"},on:{click:e.compeleteDataPlaneSetup}},[e._v(" View Your Dataplane ")])],1)]},proxy:!0},{key:"error-title",fn:function(){return[a("h3",[e._v("Mesh not found")])]},proxy:!0},{key:"error-content",fn:function(){return[a("p",[e._v("We were unable to find your mesh.")])]},proxy:!0}],null,!1,2302604054)})],1):a("KAlert",{attrs:{appearance:"danger"},scopedSlots:e._u([{key:"alertMessage",fn:function(){return[a("p",[e._v(" Please return to the first step and make sure to select an existing Mesh, or create a new one. ")])]},proxy:!0}])})]},proxy:!0},{key:"dataplane",fn:function(){return[a("h3",[e._v("Dataplane")]),a("p",[e._v(" In "+e._s(e.title)+", a Dataplane resource represents a data plane proxy running alongside one of your services. Data plane proxies can be added in any Mesh that you may have created, and in Kubernetes, they will be auto-injected by "+e._s(e.title)+". ")])]},proxy:!0},{key:"example",fn:function(){return[a("h3",[e._v("Example")]),a("p",[e._v(" Below is an example of a Dataplane resource output: ")]),a("code",[a("pre",[e._v("apiVersion: 'kuma.io/v1alpha1'\nkind: Dataplane\nmesh: default\nmetadata:\n  name: dp-echo-1\n  annotations:\n    kuma.io/sidecar-injection: enabled\n    kuma.io/mesh: default\nnetworking:\n  address: 10.0.0.1\n  inbound:\n  - port: 10000\n    servicePort: 9000\n    tags:\n      kuma.io/service: echo")])])]},proxy:!0},{key:"switch",fn:function(){return[a("Switcher")]},proxy:!0}])})],1)])},s=[],r=(a("b0c0"),a("d3b7"),a("f3f3")),i=a("2f62"),o=a("0f82"),l=a("bc1e"),c=a("1373"),d=a("2791"),u=a("251b"),p=a("4c4d"),v=a("e108"),m=a("12d5"),h=a("c3b5"),y=a("b9af"),k=a.n(y),g=a("c6ec"),f={name:"DataplaneWizardKubernetes",metaInfo:{title:"Create a new Dataplane on Kubernetes"},components:{FormFragment:d["a"],Tabs:u["a"],StepSkeleton:p["a"],Switcher:v["a"],CodeView:m["a"],Scanner:h["a"]},mixins:[c["a"]],data:function(){return{productName:g["g"],schema:k.a,steps:[{label:"General",slug:"general"},{label:"Scope Settings",slug:"scope-settings"},{label:"Install",slug:"complete"}],tabs:[{hash:"#kubernetes",title:"Kubernetes"}],sidebarContent:[{name:"dataplane"},{name:"example"},{name:"switch"}],startScanner:!1,scanFound:!1,hideScannerSiblings:!1,scanError:!1,isComplete:!1,validate:{meshName:"",k8sDataplaneType:"dataplane-type-service",k8sServices:"all-services",k8sNamespace:"",k8sNamespaceSelection:"",k8sServiceDeployment:"",k8sServiceDeploymentSelection:"",k8sIngressDeployment:"",k8sIngressDeploymentSelection:"",k8sIngressType:"",k8sIngressBrand:"kong-ingress",k8sIngressSelection:""}}},computed:Object(r["a"])(Object(r["a"])({},Object(i["c"])({title:"config/getTagline",version:"config/getVersion",environment:"config/getEnvironment",meshes:"getMeshList"})),{},{dataplaneUrl:function(){var e=this.validate;return!(!e.meshName||!e.k8sNamespaceSelection)&&{name:"dataplanes",params:{mesh:e.meshName}}},codeOutput:function(){var e=Object.assign({},this.schema),t=this.validate.k8sNamespaceSelection;if(t){e.metadata.name=t,e.metadata.namespace=t,e.metadata.annotations["kuma.io/mesh"]=this.validate.meshName;var a='" | kubectl apply -f - && kubectl delete pod --all -n '.concat(t),n=this.formatForCLI(e,a);return n}},nextDisabled:function(){var e=this.validate,t=e.k8sNamespaceSelection,a=e.meshName;return!a.length||"1"===this.$route.query.step&&!t}}),watch:{"validate.k8sNamespaceSelection":function(e){this.validate.k8sNamespaceSelection=Object(l["h"])(e)},$route:function(){var e=this.$route.query.step;1===e&&(this.validate.k8sNamespaceSelection?this.nextDisabled=!1:this.nextDisabled=!0)}},methods:{hideSiblings:function(){this.hideScannerSiblings=!0},scanForEntity:function(){var e=this,t=this.validate,a=t.meshName,n=this.validate.k8sNamespaceSelection;this.scanComplete=!1,this.scanError=!1,a&&n&&o["a"].getDataplaneFromMesh(a,n).then((function(t){t&&t.name.length>0?(e.isRunning=!0,e.scanFound=!0):e.scanError=!0})).catch((function(t){e.scanError=!0,console.error(t)})).finally((function(){e.scanComplete=!0}))},compeleteDataPlaneSetup:function(){this.$store.dispatch("updateSelectedMesh",this.validate.meshName),localStorage.setItem("selectedMesh",this.validate.meshName),this.$router.push({name:"dataplanes",params:{mesh:this.validate.meshName}})}}},b=f,_=(a("faa7"),a("2877")),w=Object(_["a"])(b,n,s,!1,null,"628b63f3",null);t["default"]=w.exports},b9af:function(e,t,a){"use strict";e.exports={apiVersion:"v1",kind:"Namespace",metadata:{name:null,namespace:null,annotations:{"kuma.io/sidecar-injection":"enabled","kuma.io/mesh":null}}}},c409:function(e,t,a){},e108:function(e,t,a){"use strict";var n=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{staticClass:"wizard-switcher"},[a("KEmptyState",{ref:"emptyState",staticClass:"my-6 wizard-empty-state",attrs:{"cta-is-hidden":"","is-error":!e.environment},scopedSlots:e._u(["kubernetes"===e.environment||"universal"===e.environment?{key:"title",fn:function(){return[e._v(" Running on "),a("span",{staticClass:"env-name"},[e._v(e._s(e.environment))])]},proxy:!0}:null,{key:"message",fn:function(){return["kubernetes"===e.environment?a("div",[e.$route.name===e.wizardRoutes.kubernetes?a("div",[a("p",[e._v(" We have detected that you are running on a "),a("strong",[e._v("Kubernetes environment")]),e._v(", and we are going to be showing you instructions for Kubernetes unless you decide to visualize the instructions for Universal. ")]),a("p",[a("KButton",{attrs:{to:{name:e.wizardRoutes.universal},appearance:"secondary"}},[e._v(" Switch to Universal instructions ")])],1)]):e.$route.name===e.wizardRoutes.universal?a("div",[a("p",[e._v(" We have detected that you are running on a "),a("strong",[e._v("Kubernetes environment")]),e._v(", but you are viewing instructions for Universal. ")]),a("p",[a("KButton",{attrs:{to:{name:e.wizardRoutes.kubernetes},appearance:"secondary"}},[e._v(" Switch back to Kubernetes instructions ")])],1)]):e._e()]):"universal"===e.environment?a("div",[e.$route.name===e.wizardRoutes.kubernetes?a("div",[a("p",[e._v(" We have detected that you are running on a "),a("strong",[e._v("Universal environment")]),e._v(", but you are viewing instructions for Kubernetes. ")]),a("p",[a("KButton",{attrs:{to:{name:e.wizardRoutes.universal},appearance:"secondary"}},[e._v(" Switch back to Universal instructions ")])],1)]):e.$route.name===e.wizardRoutes.universal?a("div",[a("p",[e._v(" We have detected that you are running on a "),a("strong",[e._v("Universal environment")]),e._v(", and we are going to be showing you instructions for Universal unless you decide to visualize the instructions for Kubernetes. ")]),a("p",[a("KButton",{attrs:{to:{name:e.wizardRoutes.kubernetes},appearance:"secondary"}},[e._v(" Switch to Kubernetes instructions ")])],1)]):e._e()]):e._e()]},proxy:!0}],null,!0)})],1)},s=[],r=a("f3f3"),i=a("2f62"),o={name:"Switcher",data:function(){return{wizardRoutes:{kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"}}},computed:Object(r["a"])(Object(r["a"])({},Object(i["c"])({environment:"config/getEnvironment"})),{},{instructionsCtaText:function(){return"universal"===this.environment?"Switch to Kubernetes instructions":"Switch to Universal instructions"},instructionsCtaRoute:function(){return"kubernetes"===this.environment?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}}})},l=o,c=(a("8b9a"),a("2877")),d=Object(c["a"])(l,n,s,!1,null,"59e6452e",null);t["a"]=d.exports},ee3d:function(e,t,a){},faa7:function(e,t,a){"use strict";a("c409")}}]);