(window["webpackJsonp"]=window["webpackJsonp"]||[]).push([["dataplanes~dataplanes-gateway~dataplanes-standard"],{"0661":function(e,t,n){},"1d10":function(e,t,n){"use strict";var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"component-frame"},[e._t("default")],2)},r=[],s={name:"FrameSkeleton"},i=s,o=(n("a948"),n("2877")),l=Object(o["a"])(i,a,r,!1,null,"666bca0e",null);t["a"]=l.exports},"23d6":function(e,t,n){"use strict";n("b91b")},"517c":function(e,t,n){"use strict";n("0661")},"536d":function(e,t,n){"use strict";n("66ad")},"62e5":function(e,t){var n;n=function(){function e(e,t){var n,a,r,s,i,o,l,c,u;null==t&&(t=""),r="",i=e.length,o=null,a=0,s=0;while(s<i){if(n=e.charAt(s),"\\"===n)r+=e.slice(s,+(s+1)+1||9e9),s++;else if("("===n)if(s<i-2)if(c=e.slice(s,+(s+2)+1||9e9),"(?:"===c)s+=2,r+=c;else if("(?<"===c){a++,s+=2,l="";while(s+1<i){if(u=e.charAt(s+1),">"===u){r+="(",s++,l.length>0&&(null==o&&(o={}),o[l]=a);break}l+=u,s++}}else r+=n,a++;else r+=n;else r+=n;s++}this.rawRegex=e,this.cleanedRegex=r,this.regex=new RegExp(this.cleanedRegex,"g"+t.replace("g","")),this.mapping=o}return e.prototype.regex=null,e.prototype.rawRegex=null,e.prototype.cleanedRegex=null,e.prototype.mapping=null,e.prototype.exec=function(e){var t,n,a,r;if(this.regex.lastIndex=0,n=this.regex.exec(e),null==n)return null;if(null!=this.mapping)for(a in r=this.mapping,r)t=r[a],n[a]=n[t];return n},e.prototype.test=function(e){return this.regex.lastIndex=0,this.regex.test(e)},e.prototype.replace=function(e,t){return this.regex.lastIndex=0,e.replace(this.regex,t)},e.prototype.replaceAll=function(e,t,n){var a;null==n&&(n=0),this.regex.lastIndex=0,a=0;while(this.regex.test(e)&&(0===n||a<n))this.regex.lastIndex=0,e=e.replace(this.regex,t),a++;return[e,a]},e}(),e.exports=n},"63b5":function(e,t,n){"use strict";var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:e._u([{key:"body",fn:function(){return[n("ul",e._l(e.warnings,(function(t){var a=t.kind,r=t.payload,s=t.index;return n("li",{key:a+"/"+s,staticClass:"mb-1"},[n("KAlert",{attrs:{appearance:"warning"},scopedSlots:e._u([{key:"alertMessage",fn:function(){return[n(e.getWarningComponent(a),{tag:"component",attrs:{payload:r}})]},proxy:!0}],null,!0)})],1)})),0)]},proxy:!0}])})},r=[],s=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("span",[e._v(" "+e._s(e.payload)+" ")])},i=[],o={name:"WarningDefault",props:{payload:{type:[String,Object],required:!0}}},l=o,c=n("2877"),u=Object(c["a"])(l,s,i,!1,null,null,null),p=u.exports,d=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("span",[e._v(" Envoy ("),n("strong",[e._v(e._s(e.payload.envoy))]),e._v(") is unsupported by the current version of Kuma DP ("),n("strong",[e._v(e._s(e.payload.kumaDp))]),e._v(") [Requirements: "),n("strong",[e._v(" "+e._s(e.payload.requirements))]),e._v("] ")])},y=[],f={name:"WarningEnvoyIncompatible",props:{payload:{type:Object,required:!0}}},m=f,v=Object(c["a"])(m,d,y,!1,null,null,null),h=v.exports,b=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("span",[e._v(" There is mismatch between versions of Kuma DP ("),n("strong",[e._v(e._s(e.payload.kumaDpVersion))]),e._v(") and the Zone CP ("),n("strong",[e._v(e._s(e.payload.zoneVersion))]),e._v(") ")])},g=[],_={name:"WarningZoneAndKumaDPVersionsIncompatible",props:{payload:{type:Object,required:!0}}},E=_,x=Object(c["a"])(E,b,g,!1,null,null,null),k=x.exports,C=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("span",[e._v(" Unsupported version of Kuma DP ("),n("strong",[e._v(e._s(e.payload.kumaDpVersion))]),e._v(") ")])},w=[],S={name:"WarningUnsupportedKumaDPVersion",props:{payload:{type:Object,required:!0}}},P=S,T=Object(c["a"])(P,C,w,!1,null,null,null),D=T.exports,O=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("span",[e._v(" There is mismatch between versions of Zone CP ("),n("strong",[e._v(e._s(e.payload.zoneCpVersion))]),e._v(") and the Global CP ("),n("strong",[e._v(e._s(e.payload.globalCpVersion))]),e._v(") ")])},A=[],j={name:"WarningZoneAndGlobalCPSVersionsIncompatible",props:{payload:{type:Object,required:!0}}},I=j,L=Object(c["a"])(I,O,A,!1,null,null,null),R=L.exports,K=n("dbf3"),N={name:"Warnings",props:{warnings:{type:Array,required:!0}},methods:{getWarningComponent:function(){var e=arguments.length>0&&void 0!==arguments[0]?arguments[0]:"";switch(e){case K["b"]:return h;case K["c"]:return D;case K["f"]:return k;case K["e"]:return R;default:return p}}}},U=N,V=Object(c["a"])(U,a,r,!1,null,null,null);t["a"]=V.exports},"66ad":function(e,t,n){},"6d8a":function(e,t,n){var a,r;r=n("62e5"),a=function(){var e;function t(){}return t.LIST_ESCAPEES=["\\","\\\\",'\\"','"',"\0","","","","","","","","\b","\t","\n","\v","\f","\r","","","","","","","","","","","","","","","","","","",(e=String.fromCharCode)(133),e(160),e(8232),e(8233)],t.LIST_ESCAPED=["\\\\",'\\"','\\"','\\"',"\\0","\\x01","\\x02","\\x03","\\x04","\\x05","\\x06","\\a","\\b","\\t","\\n","\\v","\\f","\\r","\\x0e","\\x0f","\\x10","\\x11","\\x12","\\x13","\\x14","\\x15","\\x16","\\x17","\\x18","\\x19","\\x1a","\\e","\\x1c","\\x1d","\\x1e","\\x1f","\\N","\\_","\\L","\\P"],t.MAPPING_ESCAPEES_TO_ESCAPED=function(){var e,n,a,r;for(a={},e=n=0,r=t.LIST_ESCAPEES.length;0<=r?n<r:n>r;e=0<=r?++n:--n)a[t.LIST_ESCAPEES[e]]=t.LIST_ESCAPED[e];return a}(),t.PATTERN_CHARACTERS_TO_ESCAPE=new r("[\\x00-\\x1f]|Â|Â |â¨|â©"),t.PATTERN_MAPPING_ESCAPEES=new r(t.LIST_ESCAPEES.join("|").split("\\").join("\\\\")),t.PATTERN_SINGLE_QUOTING=new r("[\\s'\":{}[\\],&*#?]|^[-?|<>=!%@`]"),t.requiresDoubleQuoting=function(e){return this.PATTERN_CHARACTERS_TO_ESCAPE.test(e)},t.escapeWithDoubleQuotes=function(e){var t;return t=this.PATTERN_MAPPING_ESCAPEES.replace(e,function(e){return function(t){return e.MAPPING_ESCAPEES_TO_ESCAPED[t]}}(this)),'"'+t+'"'},t.requiresSingleQuoting=function(e){return this.PATTERN_SINGLE_QUOTING.test(e)},t.escapeWithSingleQuotes=function(e){return"'"+e.replace(/'/g,"''")+"'"},t}(),e.exports=a},"85e6":function(e,t,n){"use strict";var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("FrameSkeleton",[n("DataOverview",{attrs:{"page-size":e.pageSize,"has-error":e.hasError,"is-loading":e.isLoading,"empty-state":e.getEmptyState(),"table-data":e.buildTableData(),"table-data-is-empty":e.tableDataIsEmpty,"show-warnings":e.tableData.data.some((function(e){return e.withWarnings})),next:e.next},on:{tableAction:e.tableAction,loadData:function(t){return e.loadData(t)}},scopedSlots:e._u([{key:"additionalControls",fn:function(){return[n("KButton",{staticClass:"add-dp-button",attrs:{appearance:"primary",size:"small",to:e.dataplaneWizardRoute},nativeOn:{click:function(t){return e.onCreateClick(t)}}},[n("span",{staticClass:"custom-control-icon"},[e._v(" + ")]),e._v(" Create data plane proxy ")]),e.$route.query.ns?n("KButton",{staticClass:"back-button",attrs:{appearance:"primary",size:"small",to:e.nsBackButtonRoute}},[n("span",{staticClass:"custom-control-icon"},[e._v(" ← ")]),e._v(" View All ")]):e._e()]},proxy:!0}])}),!1===e.isEmpty?n("Tabs",{attrs:{"has-error":e.hasError,"is-loading":e.isLoading,tabs:e.filterTabs(),"initial-tab-override":"overview"},scopedSlots:e._u([{key:"tabHeader",fn:function(){return[n("div",[e.entity.basicData?n("h3",[e._v(" DPP: "+e._s(e.entity.basicData.name)+" ")]):e._e()]),n("div",[n("EntityURLControl",{attrs:{name:e.entityName,mesh:e.entityMesh}})],1)]},proxy:!0},{key:"overview",fn:function(){return[n("LabelList",{attrs:{"is-loading":e.entityIsLoading,"is-empty":e.entityIsEmpty}},[n("div",[n("ul",e._l(e.entity.basicData,(function(t,a){return n("li",{key:a},[n("div","status"===a?[n("h4",[e._v(e._s(a))]),n("div",{staticClass:"entity-status",class:{"is-offline":"offline"===t.status.toString().toLowerCase()||!1===t.status,"is-degraded":"partially degraded"===t.status.toString().toLowerCase()||!1===t.status}},[n("span",{staticClass:"entity-status__label"},[e._v(e._s(t.status))])]),n("div",{staticClass:"reason-list"},[n("ul",e._l(t.reason,(function(t){return n("li",{key:t},[n("span",{staticClass:"entity-status__dot"}),e._v(" "+e._s(t)+" ")])})),0)])]:[n("h4",[e._v(e._s(a))]),e._v(" "+e._s(t)+" ")])])})),0)]),n("div",[n("h4",[e._v("Tags")]),n("p",e._l(e.entity.tags,(function(t,a){return n("span",{key:a,staticClass:"tag-cols"},[n("span",[e._v(" "+e._s(t.label)+": ")]),n("span",[e._v(" "+e._s(t.value)+" ")])])})),0),e.entity.versions?n("div",[n("h4",[e._v("Versions")]),n("p",e._l(e.entity.versions,(function(t,a){return n("span",{key:a,staticClass:"tag-cols"},[n("span",[e._v(" "+e._s(a)+": ")]),n("span",[e._v(" "+e._s(t)+" ")])])})),0)]):e._e()])])]},proxy:!0},{key:"insights",fn:function(){return[n("StatusInfo",{attrs:{"is-empty":0===e.subscriptionsReversed.length}},[n("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:e._u([{key:"body",fn:function(){return[n("Accordion",{attrs:{"initially-open":0}},e._l(e.subscriptionsReversed,(function(t,a){return n("AccordionItem",{key:a,scopedSlots:e._u([{key:"accordion-header",fn:function(){return[n("SubscriptionHeader",{attrs:{details:t}})]},proxy:!0},{key:"accordion-content",fn:function(){return[n("SubscriptionDetails",{attrs:{details:t,"is-discovery-subscription":""}})]},proxy:!0}],null,!0)})})),1)]},proxy:!0}],null,!1,2672485894)})],1)]},proxy:!0},{key:"dpp-policies",fn:function(){return[n("DataplanePolicies",{attrs:{mesh:e.rawEntity.mesh,"dpp-name":e.rawEntity.name}})]},proxy:!0},{key:"xds-configuration",fn:function(){return[n("XdsConfiguration",{attrs:{mesh:e.rawEntity.mesh,"dpp-name":e.rawEntity.name}})]},proxy:!0},{key:"mtls",fn:function(){return[n("LabelList",{attrs:{"is-loading":e.entityIsLoading,"is-empty":e.entityIsEmpty}},[e.entity.mtls?n("ul",e._l(e.entity.mtls,(function(t,a){return n("li",{key:a},[n("h4",[e._v(e._s(t.label))]),n("p",[e._v(" "+e._s(t.value)+" ")])])})),0):n("KAlert",{attrs:{appearance:"danger"},scopedSlots:e._u([{key:"alertMessage",fn:function(){return[e._v(" This data plane proxy does not yet have mTLS configured — "),n("a",{staticClass:"external-link",attrs:{href:"https://kuma.io/docs/"+e.version+"/documentation/security/#certificates",target:"_blank"}},[e._v(" Learn About Certificates in "+e._s(e.productName)+" ")])]},proxy:!0}],null,!1,672628330)})],1)]},proxy:!0},{key:"yaml",fn:function(){return[n("YamlView",{attrs:{"is-loading":e.entityIsLoading,"is-empty":e.entityIsEmpty,content:e.rawEntity}})]},proxy:!0},{key:"warnings",fn:function(){return[n("Warnings",{attrs:{warnings:e.warnings}})]},proxy:!0}],null,!1,2761537671)}):e._e()],1)},r=[],s=(n("4de4"),n("7db0"),n("a630"),n("d81d"),n("13d5"),n("b0c0"),n("d3b7"),n("3ca3"),n("ddb0"),n("96cf"),n("c964")),i=n("f3f3"),o=n("2f62"),l=n("0f82"),c=n("027b"),u=n("bc1e"),p=n("75bb"),d=n("dbf3"),y=n("6663"),f=n("1d10"),m=n("2778"),v=n("251b"),h=n("ff9d"),b=n("0ada"),g=n("63b5"),_=n("c6ec"),E=n("520d"),x=n("3ddf"),k=n("1d3a"),C=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("StatusInfo",{attrs:{"has-error":e.hasError,"is-loading":e.isLoading,"is-empty":!e.hasItems}},[n("KCard",{attrs:{"border-variant":"noBorder"},scopedSlots:e._u([{key:"body",fn:function(){return[n("Accordion",{attrs:{"initially-open":[],"multiple-open":""}},e._l(e.items,(function(t,a){return n("AccordionItem",{key:a,scopedSlots:e._u([{key:"accordion-header",fn:function(){return[n("div",{staticClass:"flex items-center justify-between"},[n("div",["dataplane"===t.type?n("p",{staticClass:"text-lg"},[e._v(" Dataplane ")]):e._e(),"dataplane"!==t.type?n("p",{staticClass:"text-lg"},[e._v(" "+e._s(t.service)+" ")]):e._e(),n("p",{staticClass:"subtitle"},["inbound"===t.type||"outbound"===t.type?n("span",[e._v(" "+e._s(t.type)+" "+e._s(t.name)+" ")]):"service"===t.type||"dataplane"===t.type?n("span",[e._v(" "+e._s(t.type)+" ")]):e._e(),n("KPop",{attrs:{width:"300",placement:"right",trigger:"hover"},scopedSlots:e._u([{key:"content",fn:function(){return[n("div",[e._v(" "+e._s(e.POLICY_TYPE_SUBTITLE[t.type])+" ")])]},proxy:!0}],null,!0)},[n("KIcon",{staticClass:"ml-1",attrs:{icon:"help",size:"12","view-box":"0 0 16 16"}})],1)],1)]),n("div",{staticClass:"flex flex-wrap justify-end"},e._l(t.matchedPolicies,(function(t,a){return n("KBadge",{key:a,staticClass:"mr-2 mb-2"},[e._v(" "+e._s(a)+" ")])})),1)])]},proxy:!0},{key:"accordion-content",fn:function(){return[n("div",{staticClass:"policy-wrapper"},e._l(t.matchedPolicies,(function(t,a){return n("div",{key:a,staticClass:"policy-item"},[n("h4",{staticClass:"policy-type"},[e._v(" "+e._s(e.getPolicyTitle(a))+" ")]),n("ul",e._l(t,(function(t){return n("li",{key:t.name,staticClass:"my-1",attrs:{"data-testid":"policy-name"}},[n("router-link",{attrs:{to:e.getPolicyLink(t)}},[e._v(" "+e._s(t.name)+" ")])],1)})),0)])})),0)]},proxy:!0}],null,!0)})})),1)]},proxy:!0}])})],1)},w=[],S=n("ef9d"),P={inbound:"Policies applied on incoming connection on address",outbound:"Policies applied on outgoing connection to the address",service:"Policies applied on outgoing connections to service",dataplane:"Policies applied on all incoming and outgoing connections to the selected data plane proxy"},T={name:"DataplanePolicies",components:{StatusInfo:S["a"],Accordion:E["a"],AccordionItem:x["a"]},props:{mesh:{type:String,required:!0},dppName:{type:String,required:!0}},data:function(){return{hasItems:!1,isLoading:!0,hasError:!1,policies:[],searchInput:"",POLICY_MAP:_["i"],POLICY_TYPE_SUBTITLE:P}},watch:{dppName:function(){this.fetchPolicies()}},mounted:function(){this.fetchPolicies()},methods:{getPolicyTitle:function(e){var t;return(null===(t=_["i"][e])||void 0===t?void 0:t.title)||e},getPolicyLink:function(e){var t;return{name:null===(t=_["i"][e.type])||void 0===t?void 0:t.route,query:{ns:e.name},params:{mesh:e.mesh}}},fetchPolicies:function(){var e=this;return Object(s["a"])(regeneratorRuntime.mark((function t(){var n,a,r;return regeneratorRuntime.wrap((function(t){while(1)switch(t.prev=t.next){case 0:return e.hasError=!1,e.isLoading=!0,t.prev=2,t.next=5,l["a"].getDataplanePolicies({mesh:e.mesh,dppName:e.dppName});case 5:n=t.sent,a=n.items,r=n.total,e.hasItems=r>0,e.items=a,t.next=16;break;case 12:t.prev=12,t.t0=t["catch"](2),console.error(t.t0),e.hasError=!0;case 16:return t.prev=16,e.isLoading=!1,t.finish(16);case 19:case"end":return t.stop()}}),t,null,[[2,12,16,19]])})))()}}},D=T,O=(n("517c"),n("2877")),A=Object(O["a"])(D,C,w,!1,null,"f9415e5a",null),j=A.exports,I=n("2357"),L=n("0b6d"),R=n("c8b4"),K={name:"Dataplanes",components:{Warnings:g["a"],EntityURLControl:y["a"],FrameSkeleton:f["a"],DataOverview:m["a"],Tabs:v["a"],YamlView:h["a"],LabelList:b["a"],Accordion:E["a"],AccordionItem:x["a"],SubscriptionDetails:L["a"],SubscriptionHeader:R["a"],DataplanePolicies:j,XdsConfiguration:I["a"],StatusInfo:S["a"]},props:{nsBackButtonRoute:{type:Object,default:function(){return{name:"dataplanes"}}},emptyStateMsg:{type:String,default:"There are no data plane proxies present."},dataplaneApiParams:{type:Object,default:function(){return{}}},tableHeaders:{type:Array,default:function(){return[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"},{label:"Tags",key:"tags"},{label:"Last Connected",key:"lastConnected"},{label:"Last Updated",key:"lastUpdated"},{label:"Total Updates",key:"totalUpdates"},{label:"Kuma DP version",key:"dpVersion"},{label:"Envoy version",key:"envoyVersion"},{key:"warnings",hideLabel:!0}]}},tabs:{type:Array,default:function(){return[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"DPP Insights"},{hash:"#dpp-policies",title:"Policies"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#mtls",title:"Certificate Insights"},{hash:"#yaml",title:"YAML"},{hash:"#warnings",title:"Warnings"}]}}},data:function(){return{productName:_["j"],isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,warnings:[],tableDataIsEmpty:!1,tableData:{headers:[],data:[]},subscriptionsReversed:[],entity:{},rawEntity:{},pageSize:_["g"],next:null,shownTLSTab:!1,rawData:null}},computed:Object(i["a"])(Object(i["a"])({},Object(o["c"])({environment:"config/getEnvironment",queryNamespace:"getItemQueryNamespace",multicluster:"config/getMulticlusterStatus"})),{},{dataplaneWizardRoute:function(){return"universal"===this.environment?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}},version:function(){var e=this.$store.getters.getVersion;return null!==e?e:"latest"},entityName:function(){var e,t;return(null===(e=this.entity)||void 0===e||null===(t=e.basicData)||void 0===t?void 0:t.name)||""},entityMesh:function(){var e,t;return(null===(e=this.entity)||void 0===e||null===(t=e.basicData)||void 0===t?void 0:t.mesh)||""}}),watch:{$route:function(){this.loadData()}},beforeMount:function(){this.loadData()},methods:{onCreateClick:function(){c["a"].logger.info(p["a"].CREATE_DATA_PLANE_PROXY_CLICKED)},buildEntity:function(e,t,n,a){var r=n.mTLS?Object(d["p"])(n.mTLS):null;return{basicData:e,tags:t,mtls:r,versions:a}},init:function(){this.loadData()},getEmptyState:function(){return{title:"No Data",message:this.emptyStateMsg}},filterTabs:function(){return this.warnings.length?this.tabs:this.tabs.filter((function(e){return"#warnings"!==e.hash}))},buildTableData:function(){return Object(i["a"])(Object(i["a"])({},this.tableData),{},{headers:this.tableHeaders})},compatibilityKind:function(e){return Object(d["h"])(e)},tableAction:function(e){var t=e;this.getEntity(t)},parseData:function(e){var t=this;return Object(s["a"])(regeneratorRuntime.mark((function n(){var a,r,s,i,o,l,c,p,y,f,m,v,h,b,g,_,E,x,k,C,w,S,P,T,D,O;return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:a=e.dataplane,r=void 0===a?{}:a,s=e.dataplaneInsight,i=void 0===s?{}:s,o=e.name,l=void 0===o?"":o,c=e.mesh,p=void 0===c?"":c,y=i.subscriptions,f=void 0===y?[]:y,m=Object(d["i"])(r),v=Object(d["n"])(r,i),h=v.status,b=f.reduce((function(e,t){var n=t.status,a=void 0===n?{}:n,r=t.connectTime,s=t.version,i=void 0===s?{}:s,o=a.total,l=void 0===o?{}:o,c=a.lastUpdateTime,u=l.responsesSent,p=void 0===u?"0":u,d=l.responsesRejected,y=void 0===d?"0":d,f=i.kumaDp,m=void 0===f?{}:f,v=i.envoy,h=void 0===v?{}:v,b=m.version,g=h.version,_=e.selectedTime,E=e.selectedUpdateTime,x=Date.parse(r),k=Date.parse(c);return x&&(!_||x>_)&&(_=x),k&&(!E||k>E)&&(E=k),{totalUpdates:e.totalUpdates+parseInt(p,10),totalRejectedUpdates:e.totalRejectedUpdates+parseInt(y,10),dpVersion:b||e.dpVersion,envoyVersion:g||e.envoyVersion,selectedTime:_,selectedUpdateTime:E,version:i||e.version}}),{totalUpdates:0,totalRejectedUpdates:0,dpVersion:"-",envoyVersion:"-",selectedTime:NaN,selectedUpdateTime:NaN,version:{}}),g=b.totalUpdates,_=b.totalRejectedUpdates,E=b.dpVersion,x=b.envoyVersion,k=b.selectedTime,C=b.selectedUpdateTime,w=b.version,S={name:l,mesh:p,tags:m,status:h,totalUpdates:g,totalRejectedUpdates:_,dpVersion:E,envoyVersion:x,withWarnings:!1,unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:C?Object(u["f"])(new Date(C).toUTCString()):"never",lastConnected:k?Object(u["f"])(new Date(k).toUTCString()):"never",type:Object(d["l"])(r)},P=t.compatibilityKind(w),T=P.kind,n.t0=T,n.next=n.t0===d["b"]?11:n.t0===d["c"]?14:17;break;case 11:return S.unsupportedEnvoyVersion=!0,S.withWarnings=!0,n.abrupt("break",17);case 14:return S.unsupportedKumaDPVersion=!0,S.withWarnings=!0,n.abrupt("break",17);case 17:if(!t.multicluster){n.next=23;break}return n.next=20,Object(d["g"])(m,E);case 20:D=n.sent,O=D.compatible,O||(S.withWarnings=!0,S.kumaDpAndKumaCpMismatch=!0);case 23:return n.abrupt("return",S);case 24:case"end":return n.stop()}}),n)})))()},loadData:function(){var e=arguments,t=this;return Object(s["a"])(regeneratorRuntime.mark((function n(){var a,r,s,o,c,u,p;return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:return a=e.length>0&&void 0!==e[0]?e[0]:"0",t.isLoading=!0,r=t.$route.params.mesh||null,s=t.$route.query.ns||null,n.prev=4,n.next=7,Object(k["a"])({getSingleEntity:l["a"].getDataplaneOverviewFromMesh.bind(l["a"]),getAllEntities:l["a"].getAllDataplaneOverviews.bind(l["a"]),getAllEntitiesFromMesh:l["a"].getAllDataplaneOverviewsFromMesh.bind(l["a"]),size:t.pageSize,offset:a,mesh:r,query:s,params:Object(i["a"])({},t.dataplaneApiParams)});case 7:if(o=n.sent,c=o.data,u=o.next,!c.length){n.next=22;break}return t.next=u,t.rawData=c,t.getEntity({name:c[0].name}),n.next=16,Promise.all(c.map((function(e){return t.parseData(e)})));case 16:p=n.sent,t.tableData.data=p,t.tableDataIsEmpty=!1,t.isEmpty=!1,n.next=25;break;case 22:t.tableData.data=[],t.tableDataIsEmpty=!0,t.isEmpty=!0;case 25:n.next=32;break;case 27:n.prev=27,n.t0=n["catch"](4),t.hasError=!0,t.isEmpty=!0,console.error(n.t0);case 32:return n.prev=32,t.isLoading=!1,n.finish(32);case 35:case"end":return n.stop()}}),n,null,[[4,27,32,35]])})))()},getEntity:function(e){var t=this;return Object(s["a"])(regeneratorRuntime.mark((function n(){var a,r,s,o,l,c,p,y,f,m;return regeneratorRuntime.wrap((function(n){while(1)switch(n.prev=n.next){case 0:t.entityIsLoading=!0,t.entityIsEmpty=!1,a=t.rawData.find((function(t){return t.name===e.name})),r=Object(d["j"])(a),r?(s=["type","name","mesh"],o=Object(d["k"])(a)||{},l=Object(d["n"])(r,o),c=Object(d["i"])(r),p=Object(d["o"])(o),y=Object(i["a"])(Object(i["a"])({},Object(u["d"])(r,s)),{},{status:l}),t.entity=t.buildEntity(y,c,o,p),t.warnings=[],f=o.subscriptions,m=void 0===f?[]:f,t.subscriptionsReversed=Array.from(m).reverse(),m.length&&t.setEntityWarnings(m,c),t.rawEntity=Object(u["j"])(r)):(t.entity={},t.entityIsEmpty=!0),t.entityIsLoading=!1;case 6:case"end":return n.stop()}}),n)})))()},setEntityWarnings:function(e,t){var n=this;return Object(s["a"])(regeneratorRuntime.mark((function a(){var r,s,i,o,l,c,u,p,y,f,m;return regeneratorRuntime.wrap((function(a){while(1)switch(a.prev=a.next){case 0:if(r=e[e.length-1].version,s=void 0===r?{}:r,i=s.kumaDp,o=void 0===i?{}:i,l=s.envoy,c=void 0===l?{}:l,o&&c&&(u=n.compatibilityKind(s),p=u.kind,p!==d["a"]&&p!==d["d"]&&n.warnings.push(u)),!n.multicluster){a.next=10;break}return a.next=6,Object(d["g"])(t,o.version);case 6:y=a.sent,f=y.compatible,m=y.payload,f||n.warnings.push({kind:d["f"],payload:m});case 10:case"end":return a.stop()}}),a)})))()}}},N=K,U=(n("f1a1"),Object(O["a"])(N,a,r,!1,null,"6a930e22",null));t["a"]=U.exports},a948:function(e,t,n){"use strict";n("f9f3")},b6fd:function(e,t,n){},b91b:function(e,t,n){},e80b:function(e,t,n){var a=n("6d8a"),r="  ";function s(e){var t=typeof e;return e instanceof Array?"array":"string"==t?"string":"boolean"==t?"boolean":"number"==t?"number":"undefined"==t||null===e?"null":"hash"}function i(e,t){var n=s(e);switch(n){case"array":o(e,t);break;case"hash":l(e,t);break;case"string":u(e,t);break;case"null":t.push("null");break;case"number":t.push(e.toString());break;case"boolean":t.push(e?"true":"false");break}}function o(e,t){for(var n=0;n<e.length;n++){var a=e[n],s=[];i(a,s);for(var o=0;o<s.length;o++)t.push((0==o?"- ":r)+s[o])}}function l(e,t){for(var n in e){var a=[];if(e.hasOwnProperty(n)){var o=e[n];i(o,a);var l=s(o);if("string"==l||"null"==l||"number"==l||"boolean"==l)t.push(c(n)+": "+a[0]);else{t.push(c(n)+": ");for(var u=0;u<a.length;u++)t.push(r+a[u])}}}}function c(e){return e.match(/^[\w]+$/)?e:a.requiresDoubleQuoting(e)?a.escapeWithDoubleQuotes(e):a.requiresSingleQuoting(e)?a.escapeWithSingleQuotes(e):e}function u(e,t){t.push(c(e))}var p=function(e){"string"==typeof e&&(e=JSON.parse(e));var t=[];return i(e,t),t.join("\n")};e.exports=p},f1a1:function(e,t,n){"use strict";n("b6fd")},f9f3:function(e,t,n){},ff9d:function(e,t,n){"use strict";var a=function(){var e=this,t=e.$createElement,n=e._self._c||t;return n("div",{staticClass:"yaml-view"},[e.isReady?n("div",{staticClass:"yaml-view-content"},[e.isLoading||e.isEmpty?e._e():n("KCard",{attrs:{title:e.yamlTitle,"border-variant":"noBorder"},scopedSlots:e._u([{key:"body",fn:function(){return[n("KTabs",{key:e.environment,attrs:{tabs:e.tabs},scopedSlots:e._u([{key:"universal",fn:function(){return[n("KClipboardProvider",{scopedSlots:e._u([{key:"default",fn:function(t){var a=t.copyToClipboard;return[n("KPop",{attrs:{placement:"bottom"},scopedSlots:e._u([{key:"content",fn:function(){return[n("div",[n("p",[e._v("Entity copied to clipboard!")])])]},proxy:!0}],null,!0)},[n("KButton",{staticClass:"copy-button",attrs:{appearance:"primary",size:"small"},on:{click:function(){a(e.yamlContent.universal)}}},[e._v(" Copy Universal YAML ")])],1)]}}],null,!1,1536634960)}),n("Prism",{staticClass:"code-block",attrs:{language:"yaml",code:e.yamlContent.universal}})]},proxy:!0},{key:"kubernetes",fn:function(){return[n("KClipboardProvider",{scopedSlots:e._u([{key:"default",fn:function(t){var a=t.copyToClipboard;return[n("KPop",{attrs:{placement:"bottom"},scopedSlots:e._u([{key:"content",fn:function(){return[n("div",[n("p",[e._v("Entity copied to clipboard!")])])]},proxy:!0}],null,!0)},[n("KButton",{staticClass:"copy-button",attrs:{appearance:"primary",size:"small"},on:{click:function(){a(e.yamlContent.kubernetes)}}},[e._v(" Copy Kubernetes YAML ")])],1)]}}],null,!1,2265429040)}),n("Prism",{staticClass:"code-block",attrs:{language:"yaml",code:e.yamlContent.kubernetes}})]},proxy:!0}],null,!1,1506056494),model:{value:e.activeTab.hash,callback:function(t){e.$set(e.activeTab,"hash",t)},expression:"activeTab.hash"}})]},proxy:!0}],null,!1,137880475)})],1):e._e(),!0===e.loaders?n("div",[e.isLoading?n("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:e._u([{key:"title",fn:function(){return[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{attrs:{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"}})],1),e._v(" Data Loading... ")]},proxy:!0}],null,!1,3263214496)}):e._e(),e.isEmpty&&!e.isLoading?n("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:e._u([{key:"title",fn:function(){return[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),e._v(" There is no data to display. ")]},proxy:!0}],null,!1,1612658095)}):e._e(),e.hasError?n("KEmptyState",{attrs:{"cta-is-hidden":""},scopedSlots:e._u([{key:"title",fn:function(){return[n("div",{staticClass:"card-icon mb-3"},[n("KIcon",{staticClass:"kong-icon--centered",attrs:{color:"var(--yellow-200)",icon:"warning","secondary-color":"var(--black-75)",size:"42"}})],1),e._v(" An error has occurred while trying to load this data. ")]},proxy:!0}],null,!1,822917942)}):e._e()],1):e._e()])},r=[],s=(n("caad"),n("a15b"),n("b0c0"),n("4fad"),n("ac1f"),n("2532"),n("1276"),n("f3f3")),i=n("2f62"),o=n("2ccf"),l=n.n(o),c=n("e80b"),u=n.n(c),p={name:"YamlView",components:{Prism:l.a},props:{title:{type:String,default:null},content:{type:Object,default:null},loaders:{type:Boolean,default:!0},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}},data:function(){return{tabs:[{hash:"#universal",title:"Universal"},{hash:"#kubernetes",title:"Kubernetes"}]}},computed:Object(s["a"])(Object(s["a"])({},Object(i["c"])({environment:"config/getEnvironment"})),{},{isReady:function(){return!this.isEmpty&&!this.hasError&&!this.isLoading},activeTab:{get:function(){var e=this.environment;return{hash:"#".concat(e),nohash:e}},set:function(e){return{hash:"#".concat(e),nohash:e}}},yamlTitle:function(){var e;return this.title?this.title:null!==(e=this.content)&&void 0!==e&&e.name?"Entity Overview for ".concat(this.content.name):"Entity Overview"},yamlContent:function(){var e=this,t=this.content,n=function(){var t={},n=Object.assign({},e.content),a=n.name,r=n.mesh,s=n.type,i=function(){var t=Object.assign({},e.content);return delete t.type,delete t.mesh,delete t.name,!!(t&&Object.entries(t).length>0)&&t};if(t.apiVersion="kuma.io/v1alpha1",t.kind=s,void 0!==r&&(t.mesh=n.mesh),null!==a&&void 0!==a&&a.includes(".")){var o=a.split("."),l=o.pop(),c=o.join(".");t.metadata={name:c,namespace:l}}else t.metadata={name:a};return i()&&(t.spec=i()),t},a={universal:u()(t),kubernetes:u()(n())};return a}})},d=p,y=(n("23d6"),n("536d"),n("2877")),f=Object(y["a"])(d,a,r,!1,null,"78c7b522",null);t["a"]=f.exports}}]);