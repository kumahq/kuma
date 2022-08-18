"use strict";(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[134],{18299:function(t,e,a){a.d(e,{Z:function(){return O}});var s=function(){var t=this,e=t._self._c;return e("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[e("ul",t._l(t.warnings,(function({kind:a,payload:s,index:n}){return e("li",{key:`${a}/${n}`,staticClass:"mb-1"},[e("KAlert",{attrs:{appearance:"warning"},scopedSlots:t._u([{key:"alertMessage",fn:function(){return[e(t.getWarningComponent(a),{tag:"component",attrs:{payload:s}})]},proxy:!0}],null,!0)})],1)})),0)]},proxy:!0}])})},n=[],i=function(){var t=this,e=t._self._c;return e("span",[t._v(" "+t._s(t.payload)+" ")])},r=[],o={name:"WarningDefault",props:{payload:{type:[String,Object],required:!0}}},l=o,p=a(1001),u=(0,p.Z)(l,i,r,!1,null,null,null),c=u.exports,d=function(){var t=this,e=t._self._c;return e("span",[t._v(" Envoy ("),e("strong",[t._v(t._s(t.payload.envoy))]),t._v(") is unsupported by the current version of Kuma DP ("),e("strong",[t._v(t._s(t.payload.kumaDp))]),t._v(") [Requirements: "),e("strong",[t._v(" "+t._s(t.payload.requirements))]),t._v("] ")])},y=[],m={name:"WarningEnvoyIncompatible",props:{payload:{type:Object,required:!0}}},h=m,v=(0,p.Z)(h,d,y,!1,null,null,null),g=v.exports,f=function(){var t=this,e=t._self._c;return e("span",[t._v(" There is mismatch between versions of Kuma DP ("),e("strong",[t._v(t._s(t.payload.kumaDpVersion))]),t._v(") and the Zone CP ("),e("strong",[t._v(t._s(t.payload.zoneVersion))]),t._v(") ")])},_=[],b={name:"WarningZoneAndKumaDPVersionsIncompatible",props:{payload:{type:Object,required:!0}}},k=b,D=(0,p.Z)(k,f,_,!1,null,null,null),w=D.exports,C=function(){var t=this,e=t._self._c;return e("span",[t._v(" Unsupported version of Kuma DP ("),e("strong",[t._v(t._s(t.payload.kumaDpVersion))]),t._v(") ")])},E=[],S={name:"WarningUnsupportedKumaDPVersion",props:{payload:{type:Object,required:!0}}},x=S,T=(0,p.Z)(x,C,E,!1,null,null,null),I=T.exports,L=function(){var t=this,e=t._self._c;return e("span",[t._v(" There is mismatch between versions of Zone CP ("),e("strong",[t._v(t._s(t.payload.zoneCpVersion))]),t._v(") and the Global CP ("),e("strong",[t._v(t._s(t.payload.globalCpVersion))]),t._v(") ")])},P=[],Z={name:"WarningZoneAndGlobalCPSVersionsIncompatible",props:{payload:{type:Object,required:!0}}},V=Z,A=(0,p.Z)(V,L,P,!1,null,null,null),U=A.exports,W=a(65404),K={name:"WarningsWidget",props:{warnings:{type:Array,required:!0}},methods:{getWarningComponent(t=""){switch(t){case W.Bd:return g;case W.ZM:return I;case W.pC:return w;case W.s9:return U;default:return c}}}},R=K,N=(0,p.Z)(R,s,n,!1,null,null,null),O=N.exports},11134:function(t,e,a){a.d(e,{Z:function(){return N}});var s=function(){var t=this,e=t._self._c;return e("FrameSkeleton",[e("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.getEmptyState(),"table-data":t.buildTableData(),"table-data-is-empty":t.tableDataIsEmpty,"show-warnings":t.tableData.data.some((t=>t.withWarnings)),next:t.next},on:{tableAction:t.tableAction,loadData:function(e){return t.loadData(e)}},scopedSlots:t._u([{key:"additionalControls",fn:function(){return[e("KButton",{staticClass:"add-dp-button",attrs:{appearance:"primary",size:"small",to:t.dataplaneWizardRoute},nativeOn:{click:function(e){return t.onCreateClick.apply(null,arguments)}}},[e("span",{staticClass:"custom-control-icon"},[t._v(" + ")]),t._v(" Create data plane proxy ")]),t.$route.query.ns?e("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:t.nsBackButtonRoute}},[e("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()]},proxy:!0}])}),!1===t.isEmpty?e("TabsWidget",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.filterTabs(),"initial-tab-override":"overview"},scopedSlots:t._u([{key:"tabHeader",fn:function(){return[e("div",[t.entity.basicData?e("h3",[t._v(" DPP: "+t._s(t.entity.basicData.name)+" ")]):t._e()]),e("div",[e("EntityURLControl",{attrs:{name:t.entityName,mesh:t.entityMesh}})],1)]},proxy:!0},{key:"overview",fn:function(){return[e("LabelList",{attrs:{"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[e("div",[e("ul",t._l(t.entity.basicData,(function(a,s){return e("li",{key:s},[e("div","status"===s?[e("h4",[t._v(t._s(s))]),e("div",{staticClass:"entity-status",class:{"is-offline":"offline"===a.status.toString().toLowerCase()||!1===a.status,"is-degraded":"partially degraded"===a.status.toString().toLowerCase()||!1===a.status}},[e("span",{staticClass:"entity-status__label"},[t._v(t._s(a.status))])]),e("div",{staticClass:"reason-list"},[e("ul",t._l(a.reason,(function(a){return e("li",{key:a},[e("span",{staticClass:"entity-status__dot"}),t._v(" "+t._s(a)+" ")])})),0)])]:[e("h4",[t._v(t._s(s))]),t._v(" "+t._s(a)+" ")])])})),0)]),e("div",[e("h4",[t._v("Tags")]),e("p",t._l(t.entity.tags,(function(a,s){return e("span",{key:s,staticClass:"tag-cols"},[e("span",[t._v(" "+t._s(a.label)+": ")]),e("span",[t._v(" "+t._s(a.value)+" ")])])})),0),t.entity.versions?e("div",[e("h4",[t._v("Versions")]),e("p",t._l(t.entity.versions,(function(a,s){return e("span",{key:s,staticClass:"tag-cols"},[e("span",[t._v(" "+t._s(s)+": ")]),e("span",[t._v(" "+t._s(a)+" ")])])})),0)]):t._e()])])]},proxy:!0},{key:"insights",fn:function(){return[e("StatusInfo",{attrs:{"is-empty":0===t.subscriptionsReversed.length}},[e("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[e("AccordionList",{attrs:{"initially-open":0}},t._l(t.subscriptionsReversed,(function(a,s){return e("AccordionItem",{key:s,scopedSlots:t._u([{key:"accordion-header",fn:function(){return[e("SubscriptionHeader",{attrs:{details:a}})]},proxy:!0},{key:"accordion-content",fn:function(){return[e("SubscriptionDetails",{attrs:{details:a,"is-discovery-subscription":""}})]},proxy:!0}],null,!0)})})),1)]},proxy:!0}],null,!1,4118320068)})],1)]},proxy:!0},{key:"dpp-policies",fn:function(){return[e("DataplanePolicies",{attrs:{mesh:t.rawEntity.mesh,"dpp-name":t.rawEntity.name}})]},proxy:!0},{key:"xds-configuration",fn:function(){return[e("XdsConfiguration",{attrs:{mesh:t.rawEntity.mesh,"dpp-name":t.rawEntity.name}})]},proxy:!0},{key:"envoy-stats",fn:function(){return[e("EnvoyStats",{attrs:{mesh:t.rawEntity.mesh,"dpp-name":t.rawEntity.name}})]},proxy:!0},{key:"envoy-clusters",fn:function(){return[e("EnvoyClusters",{attrs:{mesh:t.rawEntity.mesh,"dpp-name":t.rawEntity.name}})]},proxy:!0},{key:"mtls",fn:function(){return[e("LabelList",{attrs:{"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[t.entity.mtls?e("ul",t._l(t.entity.mtls,(function(a,s){return e("li",{key:s},[e("h4",[t._v(t._s(a.label))]),e("p",[t._v(" "+t._s(a.value)+" ")])])})),0):e("KAlert",{attrs:{appearance:"danger"},scopedSlots:t._u([{key:"alertMessage",fn:function(){return[t._v(" This data plane proxy does not yet have mTLS configured — "),e("a",{staticClass:"external-link",attrs:{href:`https://kuma.io/docs/${t.kumaDocsVersion}/documentation/security/#certificates`,target:"_blank"}},[t._v(" Learn About Certificates in "+t._s(t.productName)+" ")])]},proxy:!0}],null,!1,1875352035)})],1)]},proxy:!0},{key:"yaml",fn:function(){return[e("YamlView",{attrs:{"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})]},proxy:!0},{key:"warnings",fn:function(){return[e("WarningsWidget",{attrs:{warnings:t.warnings}})]},proxy:!0}],null,!1,2767641016)}):t._e()],1)},n=[],i=a(20629),r=a(89340),o=a(99716),l=a(4104),p=a(17463),u=a(53419),c=a(70878),d=a(65404),y=a(87673),m=a(84855),h=a(56882),v=a(7001),g=a(59316),f=a(33561),_=a(18299),b=a(45689),k=a(74473),D=a(49718),w=a(70172),C=function(){var t=this,e=t._self._c;return e("StatusInfo",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,"is-empty":!t.hasItems}},[e("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[e("AccordionList",{attrs:{"initially-open":[],"multiple-open":""}},t._l(t.items,(function(a,s){return e("AccordionItem",{key:s,scopedSlots:t._u([{key:"accordion-header",fn:function(){return[e("div",{staticClass:"flex items-center justify-between"},[e("div",["dataplane"===a.type?e("p",{staticClass:"text-lg"},[t._v(" Dataplane ")]):t._e(),"dataplane"!==a.type?e("p",{staticClass:"text-lg"},[t._v(" "+t._s(a.service)+" ")]):t._e(),e("p",{staticClass:"subtitle"},["inbound"===a.type||"outbound"===a.type?e("span",[t._v(" "+t._s(a.type)+" "+t._s(a.name)+" ")]):"service"===a.type||"dataplane"===a.type?e("span",[t._v(" "+t._s(a.type)+" ")]):t._e(),e("KPop",{attrs:{width:"300",placement:"right",trigger:"hover"},scopedSlots:t._u([{key:"content",fn:function(){return[e("div",[t._v(" "+t._s(t.POLICY_TYPE_SUBTITLE[a.type])+" ")])]},proxy:!0}],null,!0)},[e("KIcon",{staticClass:"ml-1",attrs:{icon:"help",size:"12","view-box":"0 0 16 16"}})],1)],1)]),e("div",{staticClass:"flex flex-wrap justify-end"},t._l(a.matchedPolicies,(function(a,s){return e("KBadge",{key:s,staticClass:"mr-2 mb-2"},[t._v(" "+t._s(s)+" ")])})),1)])]},proxy:!0},{key:"accordion-content",fn:function(){return[e("div",{staticClass:"policy-wrapper"},t._l(a.policyTypes,(function(a,n){return e("div",{key:`${s}-${n}`,staticClass:"policy-item"},[e("h4",{staticClass:"policy-type"},[t._v(" "+t._s(a.pluralDisplayName)+" ")]),e("ul",t._l(a.policies,(function(a,i){return e("li",{key:`${s}-${n}-${i}`,staticClass:"my-1",attrs:{"data-testid":"policy-name"}},[e("router-link",{attrs:{to:a.route}},[t._v(" "+t._s(a.name)+" ")])],1)})),0)])})),0)]},proxy:!0}],null,!0)})})),1)]},proxy:!0}])})],1)},E=[],S=a(9637);const x={inbound:"Policies applied on incoming connection on address",outbound:"Policies applied on outgoing connection to the address",service:"Policies applied on outgoing connections to service",dataplane:"Policies applied on all incoming and outgoing connections to the selected data plane proxy"};var T={name:"DataplanePolicies",components:{StatusInfo:S.Z,AccordionList:k.Z,AccordionItem:D.Z},props:{mesh:{type:String,required:!0},dppName:{type:String,required:!0}},data(){return{items:[],hasItems:!1,isLoading:!0,hasError:!1,searchInput:"",POLICY_TYPE_SUBTITLE:x}},computed:{...(0,i.rn)({policiesByType:t=>t.policiesByType})},watch:{dppName(){this.fetchPolicies()}},mounted(){this.fetchPolicies()},methods:{async fetchPolicies(){this.hasError=!1,this.isLoading=!0;try{const{items:t,total:e,kind:a}=await p.Z.getDataplanePolicies({mesh:this.mesh,dppName:this.dppName});void 0!==a&&"SidecarDataplane"!==a||(this.processItems(t),this.items=t,this.hasItems=e>0)}catch(t){console.error(t),this.hasError=!0}finally{this.isLoading=!1}},processItems(t){for(const e of t){e.policyTypes={};for(const t in e.matchedPolicies){const a=this.policiesByType[t],s={pluralDisplayName:a.pluralDisplayName,policies:e.matchedPolicies[t]};for(const t of s.policies)t.route={name:a.path,query:{ns:t.name},params:{mesh:t.mesh}};e.policyTypes[t]=s}}}}},I=T,L=a(1001),P=(0,L.Z)(I,C,E,!1,null,"0ef2005d",null),Z=P.exports,V=a(66190),A=a(64082),U=a(46077),W={name:"DataplanesView",components:{EnvoyStats:A.Z,EnvoyClusters:U.Z,WarningsWidget:_.Z,EntityURLControl:y.Z,FrameSkeleton:m.Z,DataOverview:h.Z,TabsWidget:v.Z,YamlView:g.Z,LabelList:f.Z,AccordionList:k.Z,AccordionItem:D.Z,SubscriptionDetails:o.Z,SubscriptionHeader:l.Z,DataplanePolicies:Z,XdsConfiguration:V.Z,StatusInfo:S.Z},props:{nsBackButtonRoute:{type:Object,default(){return{name:"dataplanes"}}},emptyStateMsg:{type:String,default:"There are no data plane proxies present."},dataplaneApiParams:{type:Object,default(){return{}}},tableHeaders:{type:Array,default(){return[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"},{label:"Tags",key:"tags"},{label:"Last Connected",key:"lastConnected"},{label:"Last Updated",key:"lastUpdated"},{label:"Total Updates",key:"totalUpdates"},{label:"Kuma DP version",key:"dpVersion"},{label:"Envoy version",key:"envoyVersion"},{key:"warnings",hideLabel:!0}]}},tabs:{type:Array,default(){return[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"DPP Insights"},{hash:"#dpp-policies",title:"Policies"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"},{hash:"#mtls",title:"Certificate Insights"},{hash:"#yaml",title:"YAML"},{hash:"#warnings",title:"Warnings"}]}}},data(){return{productName:b.sG,isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,warnings:[],tableDataIsEmpty:!1,tableData:{headers:[],data:[]},subscriptionsReversed:[],entity:{},rawEntity:{},pageSize:b.NR,next:null,shownTLSTab:!1,rawData:null}},computed:{...(0,i.Se)({environment:"config/getEnvironment",queryNamespace:"getItemQueryNamespace",multicluster:"config/getMulticlusterStatus"}),dataplaneWizardRoute(){return"universal"===this.environment?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}},kumaDocsVersion(){const t=this.$store.getters.getKumaDocsVersion;return null!==t?t:"latest"},entityName(){return this.entity?.basicData?.name||""},entityMesh(){return this.entity?.basicData?.mesh||""}},watch:{$route(){this.loadData()}},beforeMount(){this.loadData()},methods:{onCreateClick(){r.fy.logger.info(c.T.CREATE_DATA_PLANE_PROXY_CLICKED)},buildEntity(t,e,a,s){const n=a.mTLS?(0,d.Xj)(a.mTLS):null;return{basicData:t,tags:e,mtls:n,versions:s}},init(){this.loadData()},getEmptyState(){return{title:"No Data",message:this.emptyStateMsg}},filterTabs(){return this.warnings.length?this.tabs:this.tabs.filter((t=>"#warnings"!==t.hash))},buildTableData(){return{...this.tableData,headers:this.tableHeaders}},compatibilityKind(t){return(0,d.JD)(t)},tableAction(t){const e=t;this.getEntity(e)},async parseData(t){const{dataplane:e={},dataplaneInsight:a={}}=t,{name:s="",mesh:n=""}=t,{subscriptions:i=[]}=a,r=(0,d.wY)(e),{status:o}=(0,d.lR)(e,a),{totalUpdates:l,totalRejectedUpdates:p,dpVersion:c,envoyVersion:y,selectedTime:m,selectedUpdateTime:h,version:v}=i.reduce(((t,e)=>{const{status:a={},connectTime:s,version:n={}}=e,{total:i={},lastUpdateTime:r}=a,{responsesSent:o="0",responsesRejected:l="0"}=i,{kumaDp:p={},envoy:u={}}=n,{version:c}=p,{version:d}=u;let{selectedTime:y,selectedUpdateTime:m}=t;const h=Date.parse(s),v=Date.parse(r);return h&&(!y||h>y)&&(y=h),v&&(!m||v>m)&&(m=v),{totalUpdates:t.totalUpdates+parseInt(o,10),totalRejectedUpdates:t.totalRejectedUpdates+parseInt(l,10),dpVersion:c||t.dpVersion,envoyVersion:d||t.envoyVersion,selectedTime:y,selectedUpdateTime:m,version:n||t.version}}),{totalUpdates:0,totalRejectedUpdates:0,dpVersion:"-",envoyVersion:"-",selectedTime:NaN,selectedUpdateTime:NaN,version:{}}),g={name:s,mesh:n,tags:r,status:o,totalUpdates:l,totalRejectedUpdates:p,dpVersion:c,envoyVersion:y,withWarnings:!1,unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:h?(0,u.tV)(new Date(h).toUTCString()):"never",lastConnected:m?(0,u.tV)(new Date(m).toUTCString()):"never",type:(0,d.c1)(e)},{kind:f}=this.compatibilityKind(v);switch(f){case d.Bd:g.unsupportedEnvoyVersion=!0,g.withWarnings=!0;break;case d.ZM:g.unsupportedKumaDPVersion=!0,g.withWarnings=!0;break}if(this.multicluster){const{compatible:t}=await(0,d.nF)(r,c);t||(g.withWarnings=!0,g.kumaDpAndKumaCpMismatch=!0)}return g},async loadData(t="0"){this.isLoading=!0;const e=this.$route.params.mesh||null,a=this.$route.query.ns||null;try{const{data:s,next:n}=await(0,w.W)({getSingleEntity:p.Z.getDataplaneOverviewFromMesh.bind(p.Z),getAllEntities:p.Z.getAllDataplaneOverviews.bind(p.Z),getAllEntitiesFromMesh:p.Z.getAllDataplaneOverviewsFromMesh.bind(p.Z),size:this.pageSize,offset:t,mesh:e,query:a,params:{...this.dataplaneApiParams}});if(s.length){this.next=n,this.rawData=s,this.getEntity({name:s[0].name});const t=await Promise.all(s.map((t=>this.parseData(t))));this.tableData.data=t,this.tableDataIsEmpty=!1,this.isEmpty=!1}else this.tableData.data=[],this.tableDataIsEmpty=!0,this.isEmpty=!0}catch(s){this.hasError=!0,this.isEmpty=!0,console.error(s)}finally{this.isLoading=!1}},async getEntity(t){this.entityIsLoading=!0,this.entityIsEmpty=!1;const e=this.rawData.find((e=>e.name===t.name)),a=(0,d.Ng)(e);if(a){const t=["type","name","mesh"],s=(0,d.mq)(e)||{},n=(0,d.lR)(a,s),i=(0,d.wY)(a),r=(0,d.yQ)(s),o={...(0,u.wy)(a,t),status:n};this.entity=this.buildEntity(o,i,s,r),this.warnings=[];const{subscriptions:l=[]}=s;this.subscriptionsReversed=Array.from(l).reverse(),l.length&&this.setEntityWarnings(l,i),this.rawEntity=(0,u.RV)(a)}else this.entity={},this.entityIsEmpty=!0;this.entityIsLoading=!1},async setEntityWarnings(t,e){const{version:a={}}=t[t.length-1],{kumaDp:s={},envoy:n={}}=a;if(s&&n){const t=this.compatibilityKind(a),{kind:e}=t;e!==d.dG&&e!==d.O3&&this.warnings.push(t)}if(this.multicluster){const{compatible:t,payload:a}=await(0,d.nF)(e,s.version);t||this.warnings.push({kind:d.pC,payload:a})}}}},K=W,R=(0,L.Z)(K,s,n,!1,null,"7f4a2f36",null),N=R.exports}}]);