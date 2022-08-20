"use strict";(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[84],{8957:function(t,e,n){n.d(e,{Z:function(){return O}});n(92222);var a=function(){var t=this,e=t._self._c;return e("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[e("ul",t._l(t.warnings,(function(n){var a=n.kind,s=n.payload,r=n.index;return e("li",{key:"".concat(a,"/").concat(r),staticClass:"mb-1"},[e("KAlert",{attrs:{appearance:"warning"},scopedSlots:t._u([{key:"alertMessage",fn:function(){return[e(t.getWarningComponent(a),{tag:"component",attrs:{payload:s}})]},proxy:!0}],null,!0)})],1)})),0)]},proxy:!0}])})},s=[],r=function(){var t=this,e=t._self._c;return e("span",[t._v(" "+t._s(t.payload)+" ")])},i=[],o={name:"WarningDefault",props:{payload:{type:[String,Object],required:!0}}},l=o,u=n(1001),c=(0,u.Z)(l,r,i,!1,null,null,null),p=c.exports,d=function(){var t=this,e=t._self._c;return e("span",[t._v(" Envoy ("),e("strong",[t._v(t._s(t.payload.envoy))]),t._v(") is unsupported by the current version of Kuma DP ("),e("strong",[t._v(t._s(t.payload.kumaDp))]),t._v(") [Requirements: "),e("strong",[t._v(" "+t._s(t.payload.requirements))]),t._v("] ")])},y=[],m={name:"WarningEnvoyIncompatible",props:{payload:{type:Object,required:!0}}},v=m,f=(0,u.Z)(v,d,y,!1,null,null,null),h=f.exports,g=function(){var t=this,e=t._self._c;return e("span",[t._v(" There is mismatch between versions of Kuma DP ("),e("strong",[t._v(t._s(t.payload.kumaDpVersion))]),t._v(") and the Zone CP ("),e("strong",[t._v(t._s(t.payload.zoneVersion))]),t._v(") ")])},_=[],b={name:"WarningZoneAndKumaDPVersionsIncompatible",props:{payload:{type:Object,required:!0}}},k=b,w=(0,u.Z)(k,g,_,!1,null,null,null),D=w.exports,C=function(){var t=this,e=t._self._c;return e("span",[t._v(" Unsupported version of Kuma DP ("),e("strong",[t._v(t._s(t.payload.kumaDpVersion))]),t._v(") ")])},Z=[],E={name:"WarningUnsupportedKumaDPVersion",props:{payload:{type:Object,required:!0}}},x=E,S=(0,u.Z)(x,C,Z,!1,null,null,null),I=S.exports,T=function(){var t=this,e=t._self._c;return e("span",[t._v(" There is mismatch between versions of Zone CP ("),e("strong",[t._v(t._s(t.payload.zoneCpVersion))]),t._v(") and the Global CP ("),e("strong",[t._v(t._s(t.payload.globalCpVersion))]),t._v(") ")])},L=[],P={name:"WarningZoneAndGlobalCPSVersionsIncompatible",props:{payload:{type:Object,required:!0}}},A=P,V=(0,u.Z)(A,T,L,!1,null,null,null),U=V.exports,W=n(65404),K={name:"WarningsWidget",props:{warnings:{type:Array,required:!0}},methods:{getWarningComponent:function(){var t=arguments.length>0&&void 0!==arguments[0]?arguments[0]:"";switch(t){case W.Bd:return h;case W.ZM:return I;case W.pC:return D;case W.s9:return U;default:return p}}}},R=K,N=(0,u.Z)(R,a,s,!1,null,null,null),O=N.exports},73084:function(t,e,n){n.d(e,{Z:function(){return M}});n(41539),n(68309),n(39714);var a=function(){var t=this,e=t._self._c;return e("FrameSkeleton",[e("DataOverview",{attrs:{"page-size":t.pageSize,"has-error":t.hasError,"is-loading":t.isLoading,"empty-state":t.getEmptyState(),"table-data":t.buildTableData(),"table-data-is-empty":t.tableDataIsEmpty,"show-warnings":t.tableData.data.some((function(t){return t.withWarnings})),next:t.next},on:{tableAction:t.tableAction,loadData:function(e){return t.loadData(e)}},scopedSlots:t._u([{key:"additionalControls",fn:function(){return[e("KButton",{staticClass:"add-dp-button",attrs:{appearance:"primary",size:"small",to:t.dataplaneWizardRoute},nativeOn:{click:function(e){return t.onCreateClick.apply(null,arguments)}}},[e("span",{staticClass:"custom-control-icon"},[t._v(" + ")]),t._v(" Create data plane proxy ")]),t.$route.query.ns?e("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:t.nsBackButtonRoute}},[e("span",{staticClass:"custom-control-icon"},[t._v(" ← ")]),t._v(" View All ")]):t._e()]},proxy:!0}])}),!1===t.isEmpty?e("TabsWidget",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,tabs:t.filterTabs(),"initial-tab-override":"overview"},scopedSlots:t._u([{key:"tabHeader",fn:function(){return[e("div",[t.entity.basicData?e("h3",[t._v(" DPP: "+t._s(t.entity.basicData.name)+" ")]):t._e()]),e("div",[e("EntityURLControl",{attrs:{name:t.entityName,mesh:t.entityMesh}})],1)]},proxy:!0},{key:"overview",fn:function(){return[e("LabelList",{attrs:{"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[e("div",[e("ul",t._l(t.entity.basicData,(function(n,a){return e("li",{key:a},[e("div","status"===a?[e("h4",[t._v(t._s(a))]),e("div",{staticClass:"entity-status",class:{"is-offline":"offline"===n.status.toString().toLowerCase()||!1===n.status,"is-degraded":"partially degraded"===n.status.toString().toLowerCase()||!1===n.status}},[e("span",{staticClass:"entity-status__label"},[t._v(t._s(n.status))])]),e("div",{staticClass:"reason-list"},[e("ul",t._l(n.reason,(function(n){return e("li",{key:n},[e("span",{staticClass:"entity-status__dot"}),t._v(" "+t._s(n)+" ")])})),0)])]:[e("h4",[t._v(t._s(a))]),t._v(" "+t._s(n)+" ")])])})),0)]),e("div",[e("h4",[t._v("Tags")]),e("p",t._l(t.entity.tags,(function(n,a){return e("span",{key:a,staticClass:"tag-cols"},[e("span",[t._v(" "+t._s(n.label)+": ")]),e("span",[t._v(" "+t._s(n.value)+" ")])])})),0),t.entity.versions?e("div",[e("h4",[t._v("Versions")]),e("p",t._l(t.entity.versions,(function(n,a){return e("span",{key:a,staticClass:"tag-cols"},[e("span",[t._v(" "+t._s(a)+": ")]),e("span",[t._v(" "+t._s(n)+" ")])])})),0)]):t._e()])])]},proxy:!0},{key:"insights",fn:function(){return[e("StatusInfo",{attrs:{"is-empty":0===t.subscriptionsReversed.length}},[e("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[e("AccordionList",{attrs:{"initially-open":0}},t._l(t.subscriptionsReversed,(function(n,a){return e("AccordionItem",{key:a,scopedSlots:t._u([{key:"accordion-header",fn:function(){return[e("SubscriptionHeader",{attrs:{details:n}})]},proxy:!0},{key:"accordion-content",fn:function(){return[e("SubscriptionDetails",{attrs:{details:n,"is-discovery-subscription":""}})]},proxy:!0}],null,!0)})})),1)]},proxy:!0}],null,!1,4118320068)})],1)]},proxy:!0},{key:"dpp-policies",fn:function(){return[e("DataplanePolicies",{attrs:{mesh:t.rawEntity.mesh,"dpp-name":t.rawEntity.name}})]},proxy:!0},{key:"xds-configuration",fn:function(){return[e("XdsConfiguration",{attrs:{mesh:t.rawEntity.mesh,"dpp-name":t.rawEntity.name}})]},proxy:!0},{key:"envoy-stats",fn:function(){return[e("EnvoyStats",{attrs:{mesh:t.rawEntity.mesh,"dpp-name":t.rawEntity.name}})]},proxy:!0},{key:"envoy-clusters",fn:function(){return[e("EnvoyClusters",{attrs:{mesh:t.rawEntity.mesh,"dpp-name":t.rawEntity.name}})]},proxy:!0},{key:"mtls",fn:function(){return[e("LabelList",{attrs:{"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty}},[t.entity.mtls?e("ul",t._l(t.entity.mtls,(function(n,a){return e("li",{key:a},[e("h4",[t._v(t._s(n.label))]),e("p",[t._v(" "+t._s(n.value)+" ")])])})),0):e("KAlert",{attrs:{appearance:"danger"},scopedSlots:t._u([{key:"alertMessage",fn:function(){return[t._v(" This data plane proxy does not yet have mTLS configured — "),e("a",{staticClass:"external-link",attrs:{href:"https://kuma.io/docs/".concat(t.kumaDocsVersion,"/documentation/security/#certificates"),target:"_blank"}},[t._v(" Learn About Certificates in "+t._s(t.productName)+" ")])]},proxy:!0}],null,!1,1875352035)})],1)]},proxy:!0},{key:"yaml",fn:function(){return[e("YamlView",{attrs:{"is-loading":t.entityIsLoading,"is-empty":t.entityIsEmpty,content:t.rawEntity}})]},proxy:!0},{key:"warnings",fn:function(){return[e("WarningsWidget",{attrs:{warnings:t.warnings}})]},proxy:!0}],null,!1,2767641016)}):t._e()],1)},s=[],r=n(50124),i=n(48534),o=n(95082),l=(n(57327),n(78783),n(33948),n(21249),n(69826),n(91038),n(20629)),u=n(89340),c=n(99716),p=n(4104),d=n(17463),y=n(53419),m=n(70878),v=n(65404),f=n(87673),h=n(84855),g=n(56882),_=n(7001),b=n(59316),k=n(33561),w=n(8957),D=n(45689),C=n(74473),Z=n(49718),E=n(70172),x=(n(92222),function(){var t=this,e=t._self._c;return e("StatusInfo",{attrs:{"has-error":t.hasError,"is-loading":t.isLoading,"is-empty":!t.hasItems}},[e("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:t._u([{key:"body",fn:function(){return[e("AccordionList",{attrs:{"initially-open":[],"multiple-open":""}},t._l(t.items,(function(n,a){return e("AccordionItem",{key:a,scopedSlots:t._u([{key:"accordion-header",fn:function(){return[e("div",{staticClass:"flex items-center justify-between"},[e("div",["dataplane"===n.type?e("p",{staticClass:"text-lg"},[t._v(" Dataplane ")]):t._e(),"dataplane"!==n.type?e("p",{staticClass:"text-lg"},[t._v(" "+t._s(n.service)+" ")]):t._e(),e("p",{staticClass:"subtitle"},["inbound"===n.type||"outbound"===n.type?e("span",[t._v(" "+t._s(n.type)+" "+t._s(n.name)+" ")]):"service"===n.type||"dataplane"===n.type?e("span",[t._v(" "+t._s(n.type)+" ")]):t._e(),e("KPop",{attrs:{width:"300",placement:"right",trigger:"hover"},scopedSlots:t._u([{key:"content",fn:function(){return[e("div",[t._v(" "+t._s(t.POLICY_TYPE_SUBTITLE[n.type])+" ")])]},proxy:!0}],null,!0)},[e("KIcon",{staticClass:"ml-1",attrs:{icon:"help",size:"12","view-box":"0 0 16 16"}})],1)],1)]),e("div",{staticClass:"flex flex-wrap justify-end"},t._l(n.matchedPolicies,(function(n,a){return e("KBadge",{key:a,staticClass:"mr-2 mb-2"},[t._v(" "+t._s(a)+" ")])})),1)])]},proxy:!0},{key:"accordion-content",fn:function(){return[e("div",{staticClass:"policy-wrapper"},t._l(n.policyTypes,(function(n,s){return e("div",{key:"".concat(a,"-").concat(s),staticClass:"policy-item"},[e("h4",{staticClass:"policy-type"},[t._v(" "+t._s(n.pluralDisplayName)+" ")]),e("ul",t._l(n.policies,(function(n,r){return e("li",{key:"".concat(a,"-").concat(s,"-").concat(r),staticClass:"my-1",attrs:{"data-testid":"policy-name"}},[e("router-link",{attrs:{to:n.route}},[t._v(" "+t._s(n.name)+" ")])],1)})),0)])})),0)]},proxy:!0}],null,!0)})})),1)]},proxy:!0}])})],1)}),S=[],I=n(66347),T=n(9637),L={inbound:"Policies applied on incoming connection on address",outbound:"Policies applied on outgoing connection to the address",service:"Policies applied on outgoing connections to service",dataplane:"Policies applied on all incoming and outgoing connections to the selected data plane proxy"},P={name:"DataplanePolicies",components:{StatusInfo:T.Z,AccordionList:C.Z,AccordionItem:Z.Z},props:{mesh:{type:String,required:!0},dppName:{type:String,required:!0}},data:function(){return{items:[],hasItems:!1,isLoading:!0,hasError:!1,searchInput:"",POLICY_TYPE_SUBTITLE:L}},computed:(0,o.Z)({},(0,l.rn)({policiesByType:function(t){return t.policiesByType}})),watch:{dppName:function(){this.fetchPolicies()}},mounted:function(){this.fetchPolicies()},methods:{fetchPolicies:function(){var t=this;return(0,i.Z)((0,r.Z)().mark((function e(){var n,a,s,i;return(0,r.Z)().wrap((function(e){while(1)switch(e.prev=e.next){case 0:return t.hasError=!1,t.isLoading=!0,e.prev=2,e.next=5,d.Z.getDataplanePolicies({mesh:t.mesh,dppName:t.dppName});case 5:n=e.sent,a=n.items,s=n.total,i=n.kind,void 0!==i&&"SidecarDataplane"!==i||(t.processItems(a),t.items=a,t.hasItems=s>0),e.next=16;break;case 12:e.prev=12,e.t0=e["catch"](2),console.error(e.t0),t.hasError=!0;case 16:return e.prev=16,t.isLoading=!1,e.finish(16);case 19:case"end":return e.stop()}}),e,null,[[2,12,16,19]])})))()},processItems:function(t){var e,n=(0,I.Z)(t);try{for(n.s();!(e=n.n()).done;){var a=e.value;for(var s in a.policyTypes={},a.matchedPolicies){var r,i=this.policiesByType[s],o={pluralDisplayName:i.pluralDisplayName,policies:a.matchedPolicies[s]},l=(0,I.Z)(o.policies);try{for(l.s();!(r=l.n()).done;){var u=r.value;u.route={name:i.path,query:{ns:u.name},params:{mesh:u.mesh}}}}catch(c){l.e(c)}finally{l.f()}a.policyTypes[s]=o}}}catch(c){n.e(c)}finally{n.f()}}}},A=P,V=n(1001),U=(0,V.Z)(A,x,S,!1,null,"0ef2005d",null),W=U.exports,K=n(66190),R=n(64082),N=n(46077),O={name:"DataplanesView",components:{EnvoyStats:R.Z,EnvoyClusters:N.Z,WarningsWidget:w.Z,EntityURLControl:f.Z,FrameSkeleton:h.Z,DataOverview:g.Z,TabsWidget:_.Z,YamlView:b.Z,LabelList:k.Z,AccordionList:C.Z,AccordionItem:Z.Z,SubscriptionDetails:c.Z,SubscriptionHeader:p.Z,DataplanePolicies:W,XdsConfiguration:K.Z,StatusInfo:T.Z},props:{nsBackButtonRoute:{type:Object,default:function(){return{name:"dataplanes"}}},emptyStateMsg:{type:String,default:"There are no data plane proxies present."},dataplaneApiParams:{type:Object,default:function(){return{}}},tableHeaders:{type:Array,default:function(){return[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"},{label:"Tags",key:"tags"},{label:"Last Connected",key:"lastConnected"},{label:"Last Updated",key:"lastUpdated"},{label:"Total Updates",key:"totalUpdates"},{label:"Kuma DP version",key:"dpVersion"},{label:"Envoy version",key:"envoyVersion"},{key:"warnings",hideLabel:!0}]}},tabs:{type:Array,default:function(){return[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"DPP Insights"},{hash:"#dpp-policies",title:"Policies"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"},{hash:"#mtls",title:"Certificate Insights"},{hash:"#yaml",title:"YAML"},{hash:"#warnings",title:"Warnings"}]}}},data:function(){return{productName:D.sG,isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,warnings:[],tableDataIsEmpty:!1,tableData:{headers:[],data:[]},subscriptionsReversed:[],entity:{},rawEntity:{},pageSize:D.NR,next:null,shownTLSTab:!1,rawData:null}},computed:(0,o.Z)((0,o.Z)({},(0,l.Se)({environment:"config/getEnvironment",queryNamespace:"getItemQueryNamespace",multicluster:"config/getMulticlusterStatus"})),{},{dataplaneWizardRoute:function(){return"universal"===this.environment?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}},kumaDocsVersion:function(){var t=this.$store.getters.getKumaDocsVersion;return null!==t?t:"latest"},entityName:function(){var t,e;return(null===(t=this.entity)||void 0===t||null===(e=t.basicData)||void 0===e?void 0:e.name)||""},entityMesh:function(){var t,e;return(null===(t=this.entity)||void 0===t||null===(e=t.basicData)||void 0===e?void 0:e.mesh)||""}}),watch:{$route:function(){this.loadData()}},beforeMount:function(){this.loadData()},methods:{onCreateClick:function(){u.fy.logger.info(m.T.CREATE_DATA_PLANE_PROXY_CLICKED)},buildEntity:function(t,e,n,a){var s=n.mTLS?(0,v.Xj)(n.mTLS):null;return{basicData:t,tags:e,mtls:s,versions:a}},init:function(){this.loadData()},getEmptyState:function(){return{title:"No Data",message:this.emptyStateMsg}},filterTabs:function(){return this.warnings.length?this.tabs:this.tabs.filter((function(t){return"#warnings"!==t.hash}))},buildTableData:function(){return(0,o.Z)((0,o.Z)({},this.tableData),{},{headers:this.tableHeaders})},compatibilityKind:function(t){return(0,v.JD)(t)},tableAction:function(t){var e=t;this.getEntity(e)},parseData:function(t){var e=this;return(0,i.Z)((0,r.Z)().mark((function n(){var a,s,i,o,l,u,c,p,d,m,f,h,g,_,b,k,w,D,C,Z,E,x,S,I,T,L;return(0,r.Z)().wrap((function(n){while(1)switch(n.prev=n.next){case 0:a=t.dataplane,s=void 0===a?{}:a,i=t.dataplaneInsight,o=void 0===i?{}:i,l=t.name,u=void 0===l?"":l,c=t.mesh,p=void 0===c?"":c,d=o.subscriptions,m=void 0===d?[]:d,f=(0,v.wY)(s),h=(0,v.lR)(s,o),g=h.status,_=m.reduce((function(t,e){var n=e.status,a=void 0===n?{}:n,s=e.connectTime,r=e.version,i=void 0===r?{}:r,o=a.total,l=void 0===o?{}:o,u=a.lastUpdateTime,c=l.responsesSent,p=void 0===c?"0":c,d=l.responsesRejected,y=void 0===d?"0":d,m=i.kumaDp,v=void 0===m?{}:m,f=i.envoy,h=void 0===f?{}:f,g=v.version,_=h.version,b=t.selectedTime,k=t.selectedUpdateTime,w=Date.parse(s),D=Date.parse(u);return w&&(!b||w>b)&&(b=w),D&&(!k||D>k)&&(k=D),{totalUpdates:t.totalUpdates+parseInt(p,10),totalRejectedUpdates:t.totalRejectedUpdates+parseInt(y,10),dpVersion:g||t.dpVersion,envoyVersion:_||t.envoyVersion,selectedTime:b,selectedUpdateTime:k,version:i||t.version}}),{totalUpdates:0,totalRejectedUpdates:0,dpVersion:"-",envoyVersion:"-",selectedTime:NaN,selectedUpdateTime:NaN,version:{}}),b=_.totalUpdates,k=_.totalRejectedUpdates,w=_.dpVersion,D=_.envoyVersion,C=_.selectedTime,Z=_.selectedUpdateTime,E=_.version,x={name:u,mesh:p,tags:f,status:g,totalUpdates:b,totalRejectedUpdates:k,dpVersion:w,envoyVersion:D,withWarnings:!1,unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:Z?(0,y.tV)(new Date(Z).toUTCString()):"never",lastConnected:C?(0,y.tV)(new Date(C).toUTCString()):"never",type:(0,v.c1)(s)},S=e.compatibilityKind(E),I=S.kind,n.t0=I,n.next=n.t0===v.Bd?11:n.t0===v.ZM?14:17;break;case 11:return x.unsupportedEnvoyVersion=!0,x.withWarnings=!0,n.abrupt("break",17);case 14:return x.unsupportedKumaDPVersion=!0,x.withWarnings=!0,n.abrupt("break",17);case 17:if(!e.multicluster){n.next=23;break}return n.next=20,(0,v.nF)(f,w);case 20:T=n.sent,L=T.compatible,L||(x.withWarnings=!0,x.kumaDpAndKumaCpMismatch=!0);case 23:return n.abrupt("return",x);case 24:case"end":return n.stop()}}),n)})))()},loadData:function(){var t=arguments,e=this;return(0,i.Z)((0,r.Z)().mark((function n(){var a,s,i,l,u,c,p;return(0,r.Z)().wrap((function(n){while(1)switch(n.prev=n.next){case 0:return a=t.length>0&&void 0!==t[0]?t[0]:"0",e.isLoading=!0,s=e.$route.params.mesh||null,i=e.$route.query.ns||null,n.prev=4,n.next=7,(0,E.W)({getSingleEntity:d.Z.getDataplaneOverviewFromMesh.bind(d.Z),getAllEntities:d.Z.getAllDataplaneOverviews.bind(d.Z),getAllEntitiesFromMesh:d.Z.getAllDataplaneOverviewsFromMesh.bind(d.Z),size:e.pageSize,offset:a,mesh:s,query:i,params:(0,o.Z)({},e.dataplaneApiParams)});case 7:if(l=n.sent,u=l.data,c=l.next,!u.length){n.next=22;break}return e.next=c,e.rawData=u,e.getEntity({name:u[0].name}),n.next=16,Promise.all(u.map((function(t){return e.parseData(t)})));case 16:p=n.sent,e.tableData.data=p,e.tableDataIsEmpty=!1,e.isEmpty=!1,n.next=25;break;case 22:e.tableData.data=[],e.tableDataIsEmpty=!0,e.isEmpty=!0;case 25:n.next=32;break;case 27:n.prev=27,n.t0=n["catch"](4),e.hasError=!0,e.isEmpty=!0,console.error(n.t0);case 32:return n.prev=32,e.isLoading=!1,n.finish(32);case 35:case"end":return n.stop()}}),n,null,[[4,27,32,35]])})))()},getEntity:function(t){var e=this;return(0,i.Z)((0,r.Z)().mark((function n(){var a,s,i,l,u,c,p,d,m,f;return(0,r.Z)().wrap((function(n){while(1)switch(n.prev=n.next){case 0:e.entityIsLoading=!0,e.entityIsEmpty=!1,a=e.rawData.find((function(e){return e.name===t.name})),s=(0,v.Ng)(a),s?(i=["type","name","mesh"],l=(0,v.mq)(a)||{},u=(0,v.lR)(s,l),c=(0,v.wY)(s),p=(0,v.yQ)(l),d=(0,o.Z)((0,o.Z)({},(0,y.wy)(s,i)),{},{status:u}),e.entity=e.buildEntity(d,c,l,p),e.warnings=[],m=l.subscriptions,f=void 0===m?[]:m,e.subscriptionsReversed=Array.from(f).reverse(),f.length&&e.setEntityWarnings(f,c),e.rawEntity=(0,y.RV)(s)):(e.entity={},e.entityIsEmpty=!0),e.entityIsLoading=!1;case 6:case"end":return n.stop()}}),n)})))()},setEntityWarnings:function(t,e){var n=this;return(0,i.Z)((0,r.Z)().mark((function a(){var s,i,o,l,u,c,p,d,y,m,f;return(0,r.Z)().wrap((function(a){while(1)switch(a.prev=a.next){case 0:if(s=t[t.length-1].version,i=void 0===s?{}:s,o=i.kumaDp,l=void 0===o?{}:o,u=i.envoy,c=void 0===u?{}:u,l&&c&&(p=n.compatibilityKind(i),d=p.kind,d!==v.dG&&d!==v.O3&&n.warnings.push(p)),!n.multicluster){a.next=10;break}return a.next=6,(0,v.nF)(e,l.version);case 6:y=a.sent,m=y.compatible,f=y.payload,m||n.warnings.push({kind:v.pC,payload:f});case 10:case"end":return a.stop()}}),a)})))()}}},j=O,B=(0,V.Z)(j,a,s,!1,null,"7f4a2f36",null),M=B.exports},66347:function(t,e,n){n.d(e,{Z:function(){return s}});n(82526),n(41817),n(41539),n(32165),n(78783),n(33948),n(21703);var a=n(12780);function s(t,e){var n="undefined"!==typeof Symbol&&t[Symbol.iterator]||t["@@iterator"];if(!n){if(Array.isArray(t)||(n=(0,a.Z)(t))||e&&t&&"number"===typeof t.length){n&&(t=n);var s=0,r=function(){};return{s:r,n:function(){return s>=t.length?{done:!0}:{done:!1,value:t[s++]}},e:function(t){throw t},f:r}}throw new TypeError("Invalid attempt to iterate non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.")}var i,o=!0,l=!1;return{s:function(){n=n.call(t)},n:function(){var t=n.next();return o=t.done,t},e:function(t){l=!0,i=t},f:function(){try{o||null==n["return"]||n["return"]()}finally{if(l)throw i}}}}}}]);