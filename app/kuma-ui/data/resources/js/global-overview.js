(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["global-overview"],{1017:function(t,e,a){"use strict";var r=a("e756"),o=a.n(r);o.a},"10d5":function(t,e,a){"use strict";var r=a("b006"),o=a.n(r);o.a},2226:function(t,e,a){"use strict";a.r(e);var r=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"overview"},[a("MetricGrid",{attrs:{metrics:t.overviewMetrics}}),a("div",{staticClass:"card-wrapper card-wrapper--4-col"},[a("div",[a("CardSkeleton",{staticClass:"card-item",attrs:{"card-action-route":{name:"create-mesh"},"card-title":"Create a virtual Mesh","card-action-button-text":"Create Mesh"}},[a("template",{slot:"cardContent"},[a("p",[t._v(" We can create multiple isolated Mesh resources (i.e. per application/team/business unit). ")])])],2)],1),a("div",[a("CardSkeleton",{staticClass:"card-item",attrs:{"card-action-route":t.dataplaneWizardRoute,"card-title":"Connect data plane proxies","card-action-button-text":"Get Started"}},[a("template",{slot:"cardContent"},[a("p",[t._v(" We need a data plane proxy for each replicata of our services within a Mesh resource. ")])])],2)],1),a("div",[a("CardSkeleton",{staticClass:"card-item",attrs:{"card-action-route":"https://kuma.io/policies/","card-title":"Apply "+t.title+" policies","card-action-button-text":"Explore Policies"}},[a("template",{slot:"cardContent"},[a("p",[t._v(" We can apply "+t._s(t.$productName)+" policies to secure, observe, route and manage the Mesh and its data plane proxies. ")])])],2)],1),a("div",[a("Resources",{staticClass:"card-item"})],1)])],1)},o=[],c=(a("99af"),a("4de4"),a("4160"),a("159b"),a("5530")),n=a("2f62"),s=a("be10"),i=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"skeleton-card-wrapper"},[a("KCard",{class:{"is-centered":t.centerText},attrs:{title:t.cardTitle}},[a("template",{slot:"body"},[t._t("cardContent"),t.cardActionButtonText&&t.cardActionRoute?a("div",{staticClass:"skeleton-card__action mt-4"},[t._t("cardAction",[t.externalLink?a("a",{staticClass:"external-link-btn",attrs:{href:t.cardActionRoute,target:"_blank"}},[t._v(" "+t._s(t.cardActionButtonText)+" ")]):a("KButton",{attrs:{to:t.cardActionRoute,appearance:"primary"}},[t._v(" "+t._s(t.cardActionButtonText)+" ")])])],2):t._e()],2)],2)],1)},l=[],u=a("9410"),f=a.n(u),h=a("fb18"),m=a.n(h),d={name:"CardSkeleton",components:{KCard:f.a,KButton:m.a},props:{cardTitle:{type:String,required:!0},cardActionRoute:{type:[Object,String],required:!0},cardActionButtonText:{type:String,required:!0},centerText:{type:Boolean,default:!1},externalLink:{type:Boolean,default:!1}}},C=d,p=(a("1017"),a("2877")),v=Object(p["a"])(C,i,l,!1,null,"531623a5",null),T=v.exports,b=function(){var t=this,e=t.$createElement,a=t._self._c||e;return a("div",{staticClass:"resource-list"},[t.version?a("KCard",{attrs:{title:"Resources"}},[a("template",{slot:"body"},[a("p",[t._v(" Join the "+t._s(t.$productName)+" community and ask questions: ")]),a("ul",{staticClass:"resource-list"},t._l(t.resourceLinks,(function(e,r){return a("li",{key:r},[a("a",{attrs:{href:e.link,target:"_blank"}},[t._v(" "+t._s(e.label)+" ")])])})),0)])],2):t._e()],1)},k=[],M={computed:Object(c["a"])(Object(c["a"])({},Object(n["b"])({version:"getVersion"})),{},{resourceLinks:function(){var t=this.version,e=null!==t?t:"latest";return!!t&&[{link:"https://kuma.io/docs/".concat(e,"/"),label:"".concat("Kuma"," Documentation")},{link:"https://kuma-mesh.slack.com/",label:"".concat("Kuma"," Community Chat")},{link:"https://github.com/kumahq/kuma",label:"".concat("Kuma"," GitHub Repository")}]}})},_=M,x=(a("abc6"),Object(p["a"])(_,b,k,!1,null,"5141a0dd",null)),g=x.exports,y={name:"Overview",metaInfo:function(){return{title:this.$route.meta.title}},components:{MetricGrid:s["a"],CardSkeleton:T,Resources:g},computed:Object(c["a"])(Object(c["a"])({},Object(n["b"])({title:"getTagline",environment:"getEnvironment",selectedMesh:"getSelectedMesh",multicluster:"getMulticlusterStatus"})),{},{pageTitle:function(){var t=this.$route.meta.title,e=this.selectedMesh;return"all"===e?"".concat(t," for all Meshes"):"".concat(t," for ").concat(e)},overviewMetrics:function(){var t,e=this.selectedMesh,a=this.$store.state;t="all"===e?{meshCount:a.totalMeshCount,dataplaneCount:a.totalDataplaneCount,faultInjectionCount:a.totalFaultInjectionCount,healthCheckCount:a.totalHealthCheckCount,proxyTemplateCount:a.totalProxyTemplateCount,trafficLogCount:a.totalTrafficLogCount,trafficPermissionCount:a.totalTrafficPermissionCount,trafficRouteCount:a.totalTrafficRouteCount,trafficTraceCount:a.totalTrafficTraceCount,circuitBreakerCount:a.totalCircuitBreakerCount}:{dataplaneCount:a.totalDataplaneCountFromMesh,faultInjectionCount:a.totalFaultInjectionCountFromMesh,healthCheckCount:a.totalHealthCheckCountFromMesh,proxyTemplateCount:a.totalProxyTemplateCountFromMesh,trafficLogCount:a.totalTrafficLogCountFromMesh,trafficPermissionCount:a.totalTrafficPermissionCountFromMesh,trafficRouteCount:a.totalTrafficRouteCountFromMesh,trafficTraceCount:a.totalTrafficTraceCountFromMesh,circuitBreakerCount:a.totalCircuitBreakerCountFromMesh};var r=[{metric:"Meshes",value:t.meshCount,url:"/meshes/".concat(this.selectedMesh)},{metric:"Data Plane Proxies",value:t.dataplaneCount,url:"/".concat(this.selectedMesh,"/dataplanes")},{metric:"Circuit Breakers",value:t.circuitBreakerCount,url:"/".concat(this.selectedMesh,"/circuit-breakers")},{metric:"Fault Injections",value:t.faultInjectionCount,url:"/".concat(this.selectedMesh,"/fault-injections")},{metric:"Health Checks",value:t.healthCheckCount,url:"/".concat(this.selectedMesh,"/health-checks")},{metric:"Proxy Templates",value:t.proxyTemplateCount,url:"/".concat(this.selectedMesh,"/proxy-templates")},{metric:"Traffic Logs",value:t.trafficLogCount,url:"/".concat(this.selectedMesh,"/traffic-logs")},{metric:"Traffic Permissions",value:t.trafficPermissionCount,url:"/".concat(this.selectedMesh,"/traffic-permissions")},{metric:"Traffic Routes",value:t.trafficRouteCount,url:"/".concat(this.selectedMesh,"/traffic-routes")},{metric:"Traffic Traces",value:t.trafficTraceCount,url:"/".concat(this.selectedMesh,"/traffic-traces")}],o={metric:"Zones",value:this.multicluster?this.$store.state.totalClusters:"1",extraLabel:!this.multicluster&&"(Standalone)",url:"/zones"};return r.unshift(o),"all"!==e?r.filter((function(t,e,a){var r=t.metric;return"Meshes"!==r&&"Zones"!==r})):r},dataplaneWizardRoute:function(){return"universal"===this.environment?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}}}),watch:{selectedMesh:function(){this.init()}},beforeMount:function(){this.init()},methods:{init:function(){this.getCounts(),this.multicluster&&this.$store.dispatch("fetchTotalClusterCount")},getCounts:function(){var t,e=this,a=this.selectedMesh;"all"===a?(t=["fetchTotalClusterCount","fetchMeshTotalCount","fetchDataplaneTotalCount","fetchHealthCheckTotalCount","fetchProxyTemplateTotalCount","fetchTrafficLogTotalCount","fetchTrafficPermissionTotalCount","fetchTrafficRouteTotalCount","fetchTrafficTraceTotalCount","fetchFaultInjectionTotalCount","fetchCircuitBreakerTotalCount"],t.forEach((function(t){e.$store.dispatch(t)}))):(t=["fetchTotalClusterCount","fetchDataplaneTotalCountFromMesh","fetchHealthCheckTotalCountFromMesh","fetchProxyTemplateTotalCountFromMesh","fetchTrafficLogTotalCountFromMesh","fetchTrafficPermissionTotalCountFromMesh","fetchTrafficRouteTotalCountFromMesh","fetchTrafficTraceTotalCountFromMesh","fetchFaultInjectionTotalCountFromMesh","fetchCircuitBreakerTotalCountFromMesh"],t.forEach((function(t){e.$store.dispatch(t,a)})))}}},F=y,j=(a("23e4"),Object(p["a"])(F,r,o,!1,null,"6cbe065a",null));e["default"]=j.exports},"23e4":function(t,e,a){"use strict";var r=a("5752"),o=a.n(r);o.a},5752:function(t,e,a){},abc6:function(t,e,a){"use strict";var r=a("c76f"),o=a.n(r);o.a},b006:function(t,e,a){},be10:function(t,e,a){"use strict";var r=function(){var t=this,e=t.$createElement,a=t._self._c||e;return t.metrics?a("KCard",{staticClass:"info-grid-wrapper mb-4"},[a("template",{slot:"body"},[a("div",{staticClass:"info-grid",class:t.metricCountClass},t._l(t.metrics,(function(e,r){return null!==e.value?a("div",{key:r,staticClass:"metric",class:e.status,attrs:{"data-testid":t._f("formatTestId")(e.metric)}},[e.url?a("router-link",{staticClass:"metric-card",attrs:{to:e.url}},[a("div",{staticClass:"metric-title color-black-85 font-semibold"},[t._v(" "+t._s(e.metric)+" ")]),a("span",{staticClass:"metric-value mt-2 type-xl",class:{"has-error":r===t.hasError[r],"has-extra-label":e.extraLabel}},[t._v(" "+t._s(t._f("formatError")(t._f("formatValue")(e.value)))+" "),e.extraLabel?a("em",{staticClass:"metric-extra-label"},[t._v(" "+t._s(e.extraLabel)+" ")]):t._e()])]):a("div",{staticClass:"metric-card"},[a("span",{staticClass:"metric-title"},[t._v(" "+t._s(e.metric)+" ")]),a("span",{staticClass:"metric-value",class:{"has-error":r===t.hasError[r]}},[t._v(" "+t._s(t._f("formatError")(t._f("formatValue")(e.value)))+" ")])])],1):t._e()})),0)])],2):t._e()},o=[],c=(a("4160"),a("b64b"),a("d3b7"),a("ac1f"),a("25f0"),a("5319"),a("159b"),{name:"MetricsGrid",filters:{formatValue:function(t){return t?t.toLocaleString("en").toString():0},formatError:function(t){return"--"===t?"error calculating":t},formatTestId:function(t){return t.replace(" ","-").toLowerCase()}},props:{metrics:{type:Array,required:!0,default:function(){}}},computed:{hasError:function(){var t=this,e={};return Object.keys(this.metrics).forEach((function(a){"--"===t.metrics[a].value&&(e[a]=a)})),e},metricCountClass:function(){var t=this.metrics.length,e="metric-count--";return"".concat(e,t%3?"odd":"even")}}}),n=c,s=(a("10d5"),a("2877")),i=Object(s["a"])(n,r,o,!1,null,"677acc48",null);e["a"]=i.exports},c76f:function(t,e,a){},e756:function(t,e,a){}}]);