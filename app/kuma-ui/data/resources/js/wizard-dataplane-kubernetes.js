(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["wizard-dataplane-kubernetes"],{"0034":function(e,t,a){"use strict";var n=a("c241"),s=a.n(n);s.a},"0350":function(e,t){e.exports={apiVersion:"v1",kind:"Namespace",metadata:{name:null,namespace:null,labels:{"kuma.io/sidecar-injection":"enabled","kuma.io/mesh":null}}}},"124f":function(e,t,a){},"2fdb":function(e,t,a){"use strict";var n=a("5ca1"),s=a("d2c8"),r="includes";n(n.P+n.F*a("5147")(r),"String",{includes:function(e){return!!~s(this,e,r).indexOf(e,arguments.length>1?arguments[1]:void 0)}})},5147:function(e,t,a){var n=a("2b4c")("match");e.exports=function(e){var t=/./;try{"/./"[e](t)}catch(a){try{return t[n]=!1,!"/./"[e](t)}catch(s){}}return!0}},"590e":function(e,t,a){"use strict";a.d(t,"a",(function(){return s}));var n=a("bd86");a("6762"),a("2fdb"),a("ac6a"),a("456d");function s(e,t){return Object.keys(e).filter((function(e){return!t.includes(e)})).map((function(t){return Object.assign({},Object(n["a"])({},t,e[t]))})).reduce((function(e,t){return Object.assign(e,t)}),{})}},6762:function(e,t,a){"use strict";var n=a("5ca1"),s=a("c366")(!0);n(n.P,"Array",{includes:function(e){return s(this,e,arguments.length>1?arguments[1]:void 0)}}),a("9c6c")("includes")},"7a03":function(e,t,a){"use strict";var n=a("124f"),s=a.n(n);s.a},a481:function(e,t,a){"use strict";var n=a("cb7c"),s=a("4bf8"),r=a("9def"),i=a("4588"),o=a("0390"),l=a("5f1b"),c=Math.max,d=Math.min,u=Math.floor,p=/\$([$&`']|\d\d?|<[^>]*>)/g,v=/\$([$&`']|\d\d?)/g,m=function(e){return void 0===e?e:String(e)};a("214f")("replace",2,(function(e,t,a,h){return[function(n,s){var r=e(this),i=void 0==n?void 0:n[t];return void 0!==i?i.call(n,r,s):a.call(String(r),n,s)},function(e,t){var s=h(a,e,this,t);if(s.done)return s.value;var u=n(e),p=String(this),v="function"===typeof t;v||(t=String(t));var b=u.global;if(b){var y=u.unicode;u.lastIndex=0}var f=[];while(1){var k=l(u,p);if(null===k)break;if(f.push(k),!b)break;var w=String(k[0]);""===w&&(u.lastIndex=o(p,r(u.lastIndex),y))}for(var S="",_=0,D=0;D<f.length;D++){k=f[D];for(var O=String(k[0]),x=c(d(i(k.index),p.length),0),C=[],I=1;I<k.length;I++)C.push(m(k[I]));var j=k.groups;if(v){var N=[O].concat(C,x,p);void 0!==j&&N.push(j);var K=String(t.apply(void 0,N))}else K=g(O,p,x,C,j,t);x>=_&&(S+=p.slice(_,x)+K,_=x+O.length)}return S+p.slice(_)}];function g(e,t,n,r,i,o){var l=n+e.length,c=r.length,d=v;return void 0!==i&&(i=s(i),d=p),a.call(o,d,(function(a,s){var o;switch(s.charAt(0)){case"$":return"$";case"&":return e;case"`":return t.slice(0,n);case"'":return t.slice(l);case"<":o=i[s.slice(1,-1)];break;default:var d=+s;if(0===d)return a;if(d>c){var p=u(d/10);return 0===p?a:p<=c?void 0===r[p-1]?s.charAt(1):r[p-1]+s.charAt(1):a}o=r[d-1]}return void 0===o?"":o}))}}))},a527:function(e,t,a){"use strict";a.r(t);var n=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{staticClass:"wizard"},[a("div",{staticClass:"wizard__content"},[a("StepSkeleton",{attrs:{steps:e.steps,"advance-check":!0,"sidebar-content":e.sidebarContent,"footer-enabled":!1===e.hideScannerSiblings,"next-disabled":e.nextDisabled}},[a("template",{slot:"general"},[a("p",[e._v("\n            Welcome to the wizard to create a new Dataplane entity in "+e._s(e.title)+".\n            We will be providing you with a few steps that will get you started.\n          ")]),a("p",[e._v("\n            As you know, the Kuma GUI is read-only.\n          ")]),a("Switcher"),a("h3",[e._v("\n            To get started, please select on what Mesh you would like to add the Dataplane:\n          ")]),a("p",[e._v("\n            If you've got an existing Mesh that you would like to associate with your\n            Dataplane, you can select it below, or create a new one using our Mesh Wizard.\n          ")]),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""}},[a("template",{slot:"body"},[a("FormFragment",{attrs:{title:"Choose a Mesh","for-attr":"dp-mesh","all-inline":""}},[a("div",[a("select",{directives:[{name:"model",rawName:"v-model",value:e.validate.meshName,expression:"validate.meshName"}],staticClass:"k-input w-100",attrs:{id:"dp-mesh"},on:{change:function(t){var a=Array.prototype.filter.call(t.target.options,(function(e){return e.selected})).map((function(e){var t="_value"in e?e._value:e.value;return t}));e.$set(e.validate,"meshName",t.target.multiple?a:a[0])}}},[a("option",{attrs:{disabled:"",value:""}},[e._v("\n                      Select an existing Mesh…\n                    ")]),e._l(e.meshes.items,(function(t){return a("option",{key:t.name,domProps:{value:t.name}},[e._v("\n                      "+e._s(t.name)+"\n                    ")])}))],2)]),a("div",[a("label",{staticClass:"k-input-label mr-4"},[e._v("\n                    or\n                  ")]),a("KButton",{attrs:{to:{name:"create-mesh"},appearance:"primary"}},[e._v("\n                    Create a new Mesh\n                  ")])],1)])],1)],2)],1),a("template",{slot:"scope-settings"},[a("h3",[e._v("\n            Setup Dataplane Mode\n          ")]),a("p",[e._v("\n            You can create a data plane for a service or a data plane for an Ingress gateway.\n          ")]),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""}},[a("template",{slot:"body"},[a("FormFragment",{attrs:{"all-inline":"","equal-cols":"","hide-label-col":""}},[a("label",{attrs:{for:"service-dataplane"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sDataplaneType,expression:"validate.k8sDataplaneType"}],staticClass:"k-input",attrs:{id:"service-dataplane",type:"radio",name:"dataplane-type",value:"dataplane-type-service",checked:""},domProps:{checked:e._q(e.validate.k8sDataplaneType,"dataplane-type-service")},on:{change:function(t){return e.$set(e.validate,"k8sDataplaneType","dataplane-type-service")}}}),a("span",[e._v("\n                    Service Dataplane\n                  ")])]),a("label",{attrs:{for:"ingress-dataplane"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sDataplaneType,expression:"validate.k8sDataplaneType"}],staticClass:"k-input",attrs:{id:"ingress-dataplane",type:"radio",name:"dataplane-type",value:"dataplane-type-ingress",disabled:""},domProps:{checked:e._q(e.validate.k8sDataplaneType,"dataplane-type-ingress")},on:{change:function(t){return e.$set(e.validate,"k8sDataplaneType","dataplane-type-ingress")}}}),a("span",[e._v("\n                    Ingress Dataplane\n                  ")])])])],1)],2),"dataplane-type-service"===e.validate.k8sDataplaneType?a("div",[a("p",[e._v("\n              Should the data plane be added for an entire Namespace and all of its services,\n              or for specific individual services in any namespace?\n            ")]),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""}},[a("template",{slot:"body"},[a("FormFragment",{attrs:{"all-inline":"","equal-cols":"","hide-label-col":""}},[a("label",{attrs:{for:"k8s-services-all"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sServices,expression:"validate.k8sServices"}],staticClass:"k-input",attrs:{id:"k8s-services-all",type:"radio",name:"k8s-services",value:"all-services",checked:""},domProps:{checked:e._q(e.validate.k8sServices,"all-services")},on:{change:function(t){return e.$set(e.validate,"k8sServices","all-services")}}}),a("span",[e._v("\n                      All Services in Namespace\n                    ")])]),a("label",{attrs:{for:"k8s-services-individual"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sServices,expression:"validate.k8sServices"}],staticClass:"k-input",attrs:{id:"k8s-services-individual",type:"radio",name:"k8s-services",value:"individual-services",disabled:"disabled"},domProps:{checked:e._q(e.validate.k8sServices,"individual-services")},on:{change:function(t){return e.$set(e.validate,"k8sServices","individual-services")}}}),a("span",[e._v("\n                      Individual Services\n                    ")])])])],1)],2),"individual-services"===e.validate.k8sServices?a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""}},[a("template",{slot:"body"},[a("FormFragment",{attrs:{title:"Deployments","for-attr":"k8s-deployment-selection"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sServiceDeploymentSelection,expression:"validate.k8sServiceDeploymentSelection"}],staticClass:"k-input w-100",attrs:{id:"k8s-service-deployment-new",type:"text",placeholder:"your-new-deployment",required:""},domProps:{value:e.validate.k8sServiceDeploymentSelection},on:{input:function(t){t.target.composing||e.$set(e.validate,"k8sServiceDeploymentSelection",t.target.value)}}})])],1)],2):e._e(),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""}},[a("template",{slot:"body"},[a("FormFragment",{attrs:{title:"Namespace","for-attr":"k8s-namespace-selection"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sNamespaceSelection,expression:"validate.k8sNamespaceSelection"}],staticClass:"k-input w-100",attrs:{id:"k8s-namespace-new",type:"text",placeholder:"your-namespace",required:""},domProps:{value:e.validate.k8sNamespaceSelection},on:{input:function(t){t.target.composing||e.$set(e.validate,"k8sNamespaceSelection",t.target.value)}}})])],1)],2)],1):e._e(),"dataplane-type-ingress"===e.validate.k8sDataplaneType?a("div",[a("p",[e._v("\n              "+e._s(e.title)+" natively supports the Kong Ingress. Do you want to deploy\n              Kong or another Ingress?\n            ")]),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""}},[a("template",{slot:"body"},[a("FormFragment",{attrs:{"all-inline":"","equal-cols":"","hide-label-col":""}},[a("label",{attrs:{for:"k8s-ingress-kong"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sIngressBrand,expression:"validate.k8sIngressBrand"}],staticClass:"k-input",attrs:{id:"k8s-ingress-kong",type:"radio",name:"k8s-ingress-brand",value:"kong-ingress",checked:""},domProps:{checked:e._q(e.validate.k8sIngressBrand,"kong-ingress")},on:{change:function(t){return e.$set(e.validate,"k8sIngressBrand","kong-ingress")}}}),a("span",[e._v("\n                      Kong Ingress\n                    ")])]),a("label",{attrs:{for:"k8s-ingress-other"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sIngressBrand,expression:"validate.k8sIngressBrand"}],staticClass:"k-input",attrs:{id:"k8s-ingress-other",type:"radio",name:"k8s-ingress-brand",value:"other-ingress"},domProps:{checked:e._q(e.validate.k8sIngressBrand,"other-ingress")},on:{change:function(t){return e.$set(e.validate,"k8sIngressBrand","other-ingress")}}}),a("span",[e._v("\n                      Other Ingress\n                    ")])])])],1)],2),a("KCard",{staticClass:"my-6",attrs:{"has-shadow":""}},[a("template",{slot:"body"},[a("FormFragment",{attrs:{title:"Deployments","for-attr":"k8s-deployment-selection"}},[a("input",{directives:[{name:"model",rawName:"v-model",value:e.validate.k8sIngressDeployment,expression:"validate.k8sIngressDeployment"}],staticClass:"k-input w-100",attrs:{id:"k8s-ingress-deployment-new",type:"text",placeholder:"your-deployment",required:""},domProps:{value:e.validate.k8sIngressDeployment},on:{input:function(t){t.target.composing||e.$set(e.validate,"k8sIngressDeployment",t.target.value)}}})])],1)],2),"other-ingress"===e.validate.k8sIngressBrand?a("KAlert",{attrs:{appearance:"info"}},[a("template",{slot:"alertMessage"},[a("p",[e._v('\n                  Please go ahead and deploy the Ingress first, then restart this\n                  wizard and select "Existing Ingress".\n                ')])])],2):e._e()],1):e._e()],1),a("template",{slot:"complete"},[e.validate.meshName?a("div",[!1===e.hideScannerSiblings?a("div",[a("h3",[e._v("\n                Install a new Dataplane\n              ")]),a("p",[e._v("\n                You can now execute the following commands to automatically inject\n                the sidebar proxy in every Pod, and by doing so creating the Dataplane.\n              ")]),a("Tabs",{attrs:{loaders:!1,tabs:e.tabs,"has-border":!0,"initial-tab-override":"kubernetes"}},[a("template",{slot:"kubernetes"},[a("CodeView",{attrs:{title:"Kubernetes","copy-button-text":"Copy Command to Clipboard",lang:"bash",content:e.codeOutput}})],1)],2)],1):e._e()]):a("KAlert",{attrs:{appearance:"danger"}},[a("template",{slot:"alertMessage"},[a("p",[e._v("\n                Please return to the first step and make sure to select an\n                existing Mesh, or create a new one.\n              ")])])],2)],1),a("template",{slot:"dataplane"},[a("h3",[e._v("Dataplane")]),a("p",[e._v("\n            In "+e._s(e.title)+", a Dataplane entity represents a sidebar proxy running\n            alongside one of your services. Dataplanes can be added in any Mesh\n            that you may have created, and in Kubernetes, they will be auto-injected\n            by "+e._s(e.title)+".\n          ")])]),a("template",{slot:"example"},[a("h3",[e._v("Example")]),a("p",[e._v("\n            Below is an example of a Dataplane resource output:\n          ")]),a("code",[a("pre",[e._v("apiVersion: 'kuma.io/v1alpha1'\nkind: Dataplane\nmesh: default\nmetadata:\n  name: dp-echo-1\nnetworking:\n  address: 10.0.0.1\n  inbound:\n  - port: 10000\n    servicePort: 9000\n    tags:\n      service: echo")])])])],2)],1)])},s=[],r=(a("8e6e"),a("ac6a"),a("456d"),a("7f7f"),a("a481"),a("bd86")),i=a("2f62"),o=(a("590e"),a("cfb0")),l=a("ad2f"),c=a("2791"),d=a("251b"),u=a("4c4d"),p=a("e108"),v=a("12d5"),m=a("0350"),h=a.n(m);function g(e,t){var a=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),a.push.apply(a,n)}return a}function b(e){for(var t=1;t<arguments.length;t++){var a=null!=arguments[t]?arguments[t]:{};t%2?g(Object(a),!0).forEach((function(t){Object(r["a"])(e,t,a[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(a)):g(Object(a)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(a,t))}))}return e}var y={name:"DataplaneWizardKubernetes",metaInfo:{title:"Create a new Dataplane on Kubernetes"},components:{FormFragment:c["a"],Tabs:d["a"],StepSkeleton:u["a"],Switcher:p["a"],CodeView:v["a"]},mixins:[l["a"],o["a"]],data:function(){return{schema:h.a,steps:[{label:"General",slug:"general"},{label:"Scope Settings",slug:"scope-settings"},{label:"Install",slug:"complete"}],tabs:[{hash:"#kubernetes",title:"Kubernetes"}],sidebarContent:[{name:"dataplane"},{name:"example"}],startScanner:!1,scanFound:!1,hideScannerSiblings:!1,scanError:!1,isComplete:!1,nextDisabled:!0,validate:{meshName:"",k8sDataplaneType:"dataplane-type-service",k8sServices:"all-services",k8sNamespace:"",k8sNamespaceSelection:"",k8sServiceDeployment:"",k8sServiceDeploymentSelection:"",k8sIngressDeployment:"",k8sIngressDeploymentSelection:"",k8sIngressType:"",k8sIngressBrand:"kong-ingress",k8sIngressSelection:""},vmsg:[]}},computed:b({},Object(i["b"])({title:"getTagline",version:"getVersion",environment:"getEnvironment",formData:"getStoredWizardData",selectedTab:"getSelectedTab",meshes:"getMeshList"}),{codeOutput:function(){var e=Object.assign({},this.schema),t=this.validate.k8sNamespaceSelection;if(t){e.metadata.name=t,e.metadata.namespace=t,e.metadata.labels["kuma.io/mesh"]=this.validate.meshName;var a='" | kubectl apply -f - && kubectl delete pod --all -n '.concat(t),n=this.formatForCLI(e,a);return n}}}),watch:{validate:{handler:function(){var e=JSON.stringify(this.validate),t=this.validate.meshName;localStorage.setItem("storedFormData",e),t.length?this.nextDisabled=!1:this.nextDisabled=!0,1===this.$route.query.step&&(this.validate.k8sNamespaceSelection?this.nextDisabled=!1:this.nextDisabled=!0)},deep:!0},"validate.k8sNamespaceSelection":function(e){var t=e.replace(/[^a-zA-Z0-9 -]/g,"").replace(/\s+/g,"-").replace(/-+/g,"-").trim();this.validate.k8sNamespaceSelection=t},$route:function(){var e=this.$route.query.step;1===e&&(this.validate.k8sNamespaceSelection?this.nextDisabled=!1:this.nextDisabled=!0)}},methods:{scanForEntity:function(){var e=this,t=this.validate,a=t.meshName,n="test";this.scanComplete=!1,this.scanError=!1,a&&n&&this.$api.getDataplaneFromMesh(a,n).then((function(t){t&&t.name.length>0?(e.isRunning=!0,e.scanFound=!0):e.scanError=!0})).catch((function(t){e.scanError=!0,console.error(t)})).finally((function(){e.scanComplete=!0}))}}},f=y,k=(a("7a03"),a("2877")),w=Object(k["a"])(f,n,s,!1,null,"0c1aaec0",null);t["default"]=w.exports},c241:function(e,t,a){},d2c8:function(e,t,a){var n=a("aae3"),s=a("be13");e.exports=function(e,t,a){if(n(t))throw TypeError("String#"+a+" doesn't accept regex!");return String(s(e))}},e108:function(e,t,a){"use strict";var n=function(){var e=this,t=e.$createElement,a=e._self._c||t;return a("div",{staticClass:"wizard-switcher"},[a("KEmptyState",{staticClass:"my-6 empty-state--wide-content empty-state--compact",attrs:{"cta-is-hidden":"","is-error":!e.environment}},["kubernetes"===e.environment||"universal"===e.environment?a("template",{slot:"title"},[e._v("\n      Running on "),a("span",{staticClass:"env-name"},[e._v(e._s(e.environment))])]):e._e(),a("template",{slot:"message"},["kubernetes"===e.environment?a("div",[this.$route.name===e.wizardRoutes.kubernetes?a("div",[a("p",[e._v("\n            We have detected that you are running on a "),a("strong",[e._v("Kubernetes environment")]),e._v(",\n            and we are going to be showing you instructions for Kubernetes unless you\n            decide to visualize the instructions for Universal.\n          ")]),a("p",[a("KButton",{attrs:{to:{name:e.wizardRoutes.universal},appearance:"primary"}},[e._v("\n              Switch to Universal instructions\n            ")])],1)]):this.$route.name===e.wizardRoutes.universal?a("div",[a("p",[e._v("\n            We have detected that you are running on a "),a("strong",[e._v("Kubernetes environment")]),e._v(",\n            but you are viewing instructions for Universal.\n          ")]),a("p",[a("KButton",{attrs:{to:{name:e.wizardRoutes.kubernetes},appearance:"primary"}},[e._v("\n              Switch back to Kubernetes instructions\n            ")])],1)]):e._e()]):"universal"===e.environment?a("div",[this.$route.name===e.wizardRoutes.kubernetes?a("div",[a("p",[e._v("\n            We have detected that you are running on a "),a("strong",[e._v("Universal environment")]),e._v(",\n            but you are viewing instructions for Kubernetes.\n          ")]),a("p",[a("KButton",{attrs:{to:{name:e.wizardRoutes.universal},appearance:"primary"}},[e._v("\n              Switch back to Universal instructions\n            ")])],1)]):this.$route.name===e.wizardRoutes.universal?a("div",[a("p",[e._v("\n            We have detected that you are running on a "),a("strong",[e._v("Universal environment")]),e._v(",\n            and we are going to be showing you instructions for Universal unless you\n            decide to visualize the instructions for Kubernetes.\n          ")]),a("p",[a("KButton",{attrs:{to:{name:e.wizardRoutes.kubernetes},appearance:"primary"}},[e._v("\n              Switch to Kubernetes instructions\n            ")])],1)]):e._e()]):e._e()])],2)],1)},s=[],r=(a("8e6e"),a("ac6a"),a("456d"),a("bd86")),i=a("2f62");function o(e,t){var a=Object.keys(e);if(Object.getOwnPropertySymbols){var n=Object.getOwnPropertySymbols(e);t&&(n=n.filter((function(t){return Object.getOwnPropertyDescriptor(e,t).enumerable}))),a.push.apply(a,n)}return a}function l(e){for(var t=1;t<arguments.length;t++){var a=null!=arguments[t]?arguments[t]:{};t%2?o(Object(a),!0).forEach((function(t){Object(r["a"])(e,t,a[t])})):Object.getOwnPropertyDescriptors?Object.defineProperties(e,Object.getOwnPropertyDescriptors(a)):o(Object(a)).forEach((function(t){Object.defineProperty(e,t,Object.getOwnPropertyDescriptor(a,t))}))}return e}var c={name:"Switcher",data:function(){return{wizardRoutes:{kubernetes:"kubernetes-dataplane",universal:"universal-dataplane"}}},computed:l({},Object(i["b"])({environment:"getEnvironment"}),{instructionsCtaText:function(){return"universal"===this.environment?"Switch to Kubernetes instructions":"Switch to Universal instructions"},instructionsCtaRoute:function(){return"kubernetes"===this.environment?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}}})},d=c,u=(a("0034"),a("2877")),p=Object(u["a"])(d,n,s,!1,null,"99842dc4",null);t["a"]=p.exports}}]);