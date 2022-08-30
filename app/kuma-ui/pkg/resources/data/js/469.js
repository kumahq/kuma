(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[469],{73570:function(e,t,a){var n=a(49237),s="  ";function i(e){var t=typeof e;return e instanceof Array?"array":"string"==t?"string":"boolean"==t?"boolean":"number"==t?"number":"undefined"==t||null===e?"null":"hash"}function r(e,t){var a=i(e);switch(a){case"array":o(e,t);break;case"hash":l(e,t);break;case"string":u(e,t);break;case"null":t.push("null");break;case"number":t.push(e.toString());break;case"boolean":t.push(e?"true":"false");break}}function o(e,t){for(var a=0;a<e.length;a++){var n=e[a],i=[];r(n,i);for(var o=0;o<i.length;o++)t.push((0==o?"- ":s)+i[o])}}function l(e,t){for(var a in e){var n=[];if(e.hasOwnProperty(a)){var o=e[a];r(o,n);var l=i(o);if("string"==l||"null"==l||"number"==l||"boolean"==l)t.push(p(a)+": "+n[0]);else{t.push(p(a)+": ");for(var u=0;u<n.length;u++)t.push(s+n[u])}}}}function p(e){return e.match(/^[\w]+$/)?e:n.requiresDoubleQuoting(e)?n.escapeWithDoubleQuotes(e):n.requiresSingleQuoting(e)?n.escapeWithSingleQuotes(e):e}function u(e,t){t.push(p(e))}var c=function(e){"string"==typeof e&&(e=JSON.parse(e));var t=[];return r(e,t),t.join("\n")};e.exports=c},99997:function(e,t,a){"use strict";a.d(t,{Z:function(){return S}});var n=a(70821);const s=e=>((0,n.dD)("data-v-487d5c0d"),e=e(),(0,n.Cn)(),e),i={class:"yaml-view"},r={key:0,class:"yaml-view-content"},o=(0,n.Uk)(" Copy Universal YAML "),l=s((()=>(0,n._)("div",null,[(0,n._)("p",null,"Entity copied to clipboard!")],-1))),p=(0,n.Uk)(" Copy Kubernetes YAML "),u=s((()=>(0,n._)("div",null,[(0,n._)("p",null,"Entity copied to clipboard!")],-1))),c={key:1},d={class:"card-icon mb-3"},m=(0,n.Uk)(" Data Loading... "),y={class:"card-icon mb-3"},h=(0,n.Uk)(" There is no data to display. "),g={class:"card-icon mb-3"},w=(0,n.Uk)(" An error has occurred while trying to load this data. ");function v(e,t,a,s,v,b){const f=(0,n.up)("KButton"),k=(0,n.up)("KPop"),_=(0,n.up)("KClipboardProvider"),E=(0,n.up)("CodeBlock"),D=(0,n.up)("KTabs"),C=(0,n.up)("KCard"),S=(0,n.up)("KIcon"),T=(0,n.up)("KEmptyState");return(0,n.wg)(),(0,n.iD)("div",i,[b.isReady?((0,n.wg)(),(0,n.iD)("div",r,[a.isLoading||a.isEmpty?(0,n.kq)("",!0):((0,n.wg)(),(0,n.j4)(C,{key:0,title:b.yamlTitle,"border-variant":"noBorder"},{body:(0,n.w5)((()=>[((0,n.wg)(),(0,n.j4)(D,{key:e.environment,modelValue:b.activeTab.hash,"onUpdate:modelValue":t[0]||(t[0]=e=>b.activeTab.hash=e),tabs:v.tabs},{universal:(0,n.w5)((()=>[(0,n.Wm)(_,null,{default:(0,n.w5)((({copyToClipboard:e})=>[(0,n.Wm)(k,{placement:"bottom"},{content:(0,n.w5)((()=>[l])),default:(0,n.w5)((()=>[(0,n.Wm)(f,{class:"copy-button",appearance:"primary",size:"small",onClick:()=>{e(b.yamlContent.universal)}},{default:(0,n.w5)((()=>[o])),_:2},1032,["onClick"])])),_:2},1024)])),_:1}),(0,n.Wm)(E,{language:"yaml",code:b.yamlContent.universal},null,8,["code"])])),kubernetes:(0,n.w5)((()=>[(0,n.Wm)(_,null,{default:(0,n.w5)((({copyToClipboard:e})=>[(0,n.Wm)(k,{placement:"bottom"},{content:(0,n.w5)((()=>[u])),default:(0,n.w5)((()=>[(0,n.Wm)(f,{class:"copy-button",appearance:"primary",size:"small",onClick:()=>{e(b.yamlContent.kubernetes)}},{default:(0,n.w5)((()=>[p])),_:2},1032,["onClick"])])),_:2},1024)])),_:1}),(0,n.Wm)(E,{language:"yaml",code:b.yamlContent.kubernetes},null,8,["code"])])),_:1},8,["modelValue","tabs"]))])),_:1},8,["title"]))])):(0,n.kq)("",!0),!0===a.loaders?((0,n.wg)(),(0,n.iD)("div",c,[a.isLoading?((0,n.wg)(),(0,n.j4)(T,{key:0,"cta-is-hidden":""},{title:(0,n.w5)((()=>[(0,n._)("div",d,[(0,n.Wm)(S,{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"})]),m])),_:1})):(0,n.kq)("",!0),a.isEmpty&&!a.isLoading?((0,n.wg)(),(0,n.j4)(T,{key:1,"cta-is-hidden":""},{title:(0,n.w5)((()=>[(0,n._)("div",y,[(0,n.Wm)(S,{class:"kong-icon--centered",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"42"})]),h])),_:1})):(0,n.kq)("",!0),a.hasError?((0,n.wg)(),(0,n.j4)(T,{key:2,"cta-is-hidden":""},{title:(0,n.w5)((()=>[(0,n._)("div",g,[(0,n.Wm)(S,{class:"kong-icon--centered",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"42"})]),w])),_:1})):(0,n.kq)("",!0)])):(0,n.kq)("",!0)])}var b=a(33907),f=a(73570),k=a.n(f),_=a(21743),E={name:"YamlView",components:{CodeBlock:_.Z},props:{title:{type:String,default:null},content:{type:Object,default:null},loaders:{type:Boolean,default:!0},isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1}},data(){return{tabs:[{hash:"#universal",title:"Universal"},{hash:"#kubernetes",title:"Kubernetes"}]}},computed:{...(0,b.Se)({environment:"config/getEnvironment"}),isReady(){return!this.isEmpty&&!this.hasError&&!this.isLoading},activeTab:{get(){const e=this.environment;return{hash:`#${e}`,nohash:e}},set(e){return{hash:`#${e}`,nohash:e}}},yamlTitle(){return this.title?this.title:this.content?.name?`Entity Overview for ${this.content.name}`:"Entity Overview"},yamlContent(){const e=this.content,t=()=>{const e={},t=Object.assign({},this.content),{name:a,mesh:n,type:s}=t,i=()=>{const e=Object.assign({},this.content);return delete e.type,delete e.mesh,delete e.name,!!(e&&Object.entries(e).length>0)&&e};if(e.apiVersion="kuma.io/v1alpha1",e.kind=s,void 0!==n&&(e.mesh=t.mesh),a?.includes(".")){const t=a.split("."),n=t.pop(),s=t.join(".");e.metadata={name:s,namespace:n}}else e.metadata={name:a};return i()&&(e.spec=i()),e},a={universal:k()(e),kubernetes:k()(t())};return a}}},D=a(83744);const C=(0,D.Z)(E,[["render",v],["__scopeId","data-v-487d5c0d"]]);var S=C},51372:function(e,t,a){"use strict";a.d(t,{Z:function(){return Z}});var n=a(70821);function s(e,t,a,s,i,r){const o=(0,n.up)("KAlert"),l=(0,n.up)("KCard");return(0,n.wg)(),(0,n.j4)(l,{"border-variant":"noBorder"},{body:(0,n.w5)((()=>[(0,n._)("ul",null,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(a.warnings,(({kind:e,payload:t,index:a})=>((0,n.wg)(),(0,n.iD)("li",{key:`${e}/${a}`,class:"mb-1"},[(0,n.Wm)(o,{appearance:"warning"},{alertMessage:(0,n.w5)((()=>[((0,n.wg)(),(0,n.j4)((0,n.LL)(r.getWarningComponent(e)),{payload:t},null,8,["payload"]))])),_:2},1024)])))),128))])])),_:1})}function i(e,t,a,s,i,r){return(0,n.wg)(),(0,n.iD)("span",null,(0,n.zw)(a.payload),1)}var r={name:"WarningDefault",props:{payload:{type:[String,Object],required:!0}}},o=a(83744);const l=(0,o.Z)(r,[["render",i]]);var p=l;const u=(0,n.Uk)(" Envoy ("),c=(0,n.Uk)(") is unsupported by the current version of Kuma DP ("),d=(0,n.Uk)(") [Requirements: "),m=(0,n.Uk)("] ");function y(e,t,a,s,i,r){return(0,n.wg)(),(0,n.iD)("span",null,[u,(0,n._)("strong",null,(0,n.zw)(a.payload.envoy),1),c,(0,n._)("strong",null,(0,n.zw)(a.payload.kumaDp),1),d,(0,n._)("strong",null,(0,n.zw)(a.payload.requirements),1),m])}var h={name:"WarningEnvoyIncompatible",props:{payload:{type:Object,required:!0}}};const g=(0,o.Z)(h,[["render",y]]);var w=g;const v=(0,n.Uk)(" There is mismatch between versions of Kuma DP ("),b=(0,n.Uk)(") and the Zone CP ("),f=(0,n.Uk)(") ");function k(e,t,a,s,i,r){return(0,n.wg)(),(0,n.iD)("span",null,[v,(0,n._)("strong",null,(0,n.zw)(a.payload.kumaDpVersion),1),b,(0,n._)("strong",null,(0,n.zw)(a.payload.zoneVersion),1),f])}var _={name:"WarningZoneAndKumaDPVersionsIncompatible",props:{payload:{type:Object,required:!0}}};const E=(0,o.Z)(_,[["render",k]]);var D=E;const C=(0,n.Uk)(" Unsupported version of Kuma DP ("),S=(0,n.Uk)(") ");function T(e,t,a,s,i,r){return(0,n.wg)(),(0,n.iD)("span",null,[C,(0,n._)("strong",null,(0,n.zw)(a.payload.kumaDpVersion),1),S])}var P={name:"WarningUnsupportedKumaDPVersion",props:{payload:{type:Object,required:!0}}};const x=(0,o.Z)(P,[["render",T]]);var A=x;const I=(0,n.Uk)(" There is mismatch between versions of Zone CP ("),L=(0,n.Uk)(") and the Global CP ("),W=(0,n.Uk)(") ");function U(e,t,a,s,i,r){return(0,n.wg)(),(0,n.iD)("span",null,[I,(0,n._)("strong",null,(0,n.zw)(a.payload.zoneCpVersion),1),L,(0,n._)("strong",null,(0,n.zw)(a.payload.globalCpVersion),1),W])}var z={name:"WarningZoneAndGlobalCPSVersionsIncompatible",props:{payload:{type:Object,required:!0}}};const K=(0,o.Z)(z,[["render",U]]);var R=K,V=a(65404),j={name:"WarningsWidget",props:{warnings:{type:Array,required:!0}},methods:{getWarningComponent(e=""){switch(e){case V.Bd:return w;case V.ZM:return A;case V.pC:return D;case V.s9:return R;default:return p}}}};const N=(0,o.Z)(j,[["render",s]]);var Z=N},96469:function(e,t,a){"use strict";a.d(t,{Z:function(){return le}});var n=a(70821);const s=e=>((0,n.dD)("data-v-5bfdf844"),e=e(),(0,n.Cn)(),e),i=s((()=>(0,n._)("span",{class:"custom-control-icon"}," + ",-1))),r=(0,n.Uk)(" Create data plane proxy "),o=s((()=>(0,n._)("span",{class:"custom-control-icon"}," ← ",-1))),l=(0,n.Uk)(" View All "),p={key:0},u={key:0},c={class:"entity-status__label"},d={class:"reason-list"},m=s((()=>(0,n._)("span",{class:"entity-status__dot"},null,-1))),y={key:1},h=s((()=>(0,n._)("h4",null,"Tags",-1))),g={key:0},w=s((()=>(0,n._)("h4",null,"Versions",-1))),v={key:0},b=(0,n.Uk)(" This data plane proxy does not yet have mTLS configured — "),f=["href"];function k(e,t,a,s,k,_){const E=(0,n.up)("KButton"),D=(0,n.up)("DataOverview"),C=(0,n.up)("EntityURLControl"),S=(0,n.up)("LabelList"),T=(0,n.up)("SubscriptionHeader"),P=(0,n.up)("SubscriptionDetails"),x=(0,n.up)("AccordionItem"),A=(0,n.up)("AccordionList"),I=(0,n.up)("KCard"),L=(0,n.up)("StatusInfo"),W=(0,n.up)("DataplanePolicies"),U=(0,n.up)("XdsConfiguration"),z=(0,n.up)("EnvoyStats"),K=(0,n.up)("EnvoyClusters"),R=(0,n.up)("KAlert"),V=(0,n.up)("YamlView"),j=(0,n.up)("WarningsWidget"),N=(0,n.up)("TabsWidget"),Z=(0,n.up)("FrameSkeleton");return(0,n.wg)(),(0,n.j4)(Z,null,{default:(0,n.w5)((()=>[(0,n.Wm)(D,{"page-size":k.pageSize,"has-error":k.hasError,"is-loading":k.isLoading,"empty-state":_.getEmptyState(),"table-data":_.buildTableData(),"table-data-is-empty":k.tableDataIsEmpty,"show-warnings":k.tableData.data.some((e=>e.withWarnings)),next:k.next,onTableAction:_.tableAction,onLoadData:t[0]||(t[0]=e=>_.loadData(e))},{additionalControls:(0,n.w5)((()=>[(0,n.Wm)(E,{class:"add-dp-button",appearance:"primary",size:"small",to:_.dataplaneWizardRoute,onClick:_.onCreateClick},{default:(0,n.w5)((()=>[i,r])),_:1},8,["to","onClick"]),e.$route.query.ns?((0,n.wg)(),(0,n.j4)(E,{key:0,class:"back-button",appearance:"primary",size:"small",to:a.nsBackButtonRoute},{default:(0,n.w5)((()=>[o,l])),_:1},8,["to"])):(0,n.kq)("",!0)])),_:1},8,["page-size","has-error","is-loading","empty-state","table-data","table-data-is-empty","show-warnings","next","onTableAction"]),!1===k.isEmpty?((0,n.wg)(),(0,n.j4)(N,{key:0,"has-error":k.hasError,"is-loading":k.isLoading,tabs:_.filterTabs(),"initial-tab-override":"overview"},{tabHeader:(0,n.w5)((()=>[(0,n._)("div",null,[k.entity.basicData?((0,n.wg)(),(0,n.iD)("h3",p," DPP: "+(0,n.zw)(k.entity.basicData.name),1)):(0,n.kq)("",!0)]),(0,n._)("div",null,[(0,n.Wm)(C,{name:_.entityName,mesh:_.entityMesh},null,8,["name","mesh"])])])),overview:(0,n.w5)((()=>[(0,n.Wm)(S,{"is-loading":k.entityIsLoading,"is-empty":k.entityIsEmpty},{default:(0,n.w5)((()=>[(0,n._)("div",null,[(0,n._)("ul",null,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(k.entity.basicData,((e,t)=>((0,n.wg)(),(0,n.iD)("li",{key:t},["status"===t?((0,n.wg)(),(0,n.iD)("div",u,[(0,n._)("h4",null,(0,n.zw)(t),1),(0,n._)("div",{class:(0,n.C_)(["entity-status",{"is-offline":"offline"===e.status.toString().toLowerCase()||!1===e.status,"is-degraded":"partially degraded"===e.status.toString().toLowerCase()||!1===e.status}])},[(0,n._)("span",c,(0,n.zw)(e.status),1)],2),(0,n._)("div",d,[(0,n._)("ul",null,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(e.reason,(e=>((0,n.wg)(),(0,n.iD)("li",{key:e},[m,(0,n.Uk)(" "+(0,n.zw)(e),1)])))),128))])])])):((0,n.wg)(),(0,n.iD)("div",y,[(0,n._)("h4",null,(0,n.zw)(t),1),(0,n.Uk)(" "+(0,n.zw)(e),1)]))])))),128))])]),(0,n._)("div",null,[h,(0,n._)("p",null,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(k.entity.tags,((e,t)=>((0,n.wg)(),(0,n.iD)("span",{key:t,class:"tag-cols"},[(0,n._)("span",null,(0,n.zw)(e.label)+": ",1),(0,n._)("span",null,(0,n.zw)(e.value),1)])))),128))]),k.entity.versions?((0,n.wg)(),(0,n.iD)("div",g,[w,(0,n._)("p",null,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(k.entity.versions,((e,t)=>((0,n.wg)(),(0,n.iD)("span",{key:t,class:"tag-cols"},[(0,n._)("span",null,(0,n.zw)(t)+": ",1),(0,n._)("span",null,(0,n.zw)(e),1)])))),128))])])):(0,n.kq)("",!0)])])),_:1},8,["is-loading","is-empty"])])),insights:(0,n.w5)((()=>[(0,n.Wm)(L,{"is-empty":0===k.subscriptionsReversed.length},{default:(0,n.w5)((()=>[(0,n.Wm)(I,{"border-variant":"noBorder"},{body:(0,n.w5)((()=>[(0,n.Wm)(A,{"initially-open":0},{default:(0,n.w5)((()=>[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(k.subscriptionsReversed,((e,t)=>((0,n.wg)(),(0,n.j4)(x,{key:t},{"accordion-header":(0,n.w5)((()=>[(0,n.Wm)(T,{details:e},null,8,["details"])])),"accordion-content":(0,n.w5)((()=>[(0,n.Wm)(P,{details:e,"is-discovery-subscription":""},null,8,["details"])])),_:2},1024)))),128))])),_:1})])),_:1})])),_:1},8,["is-empty"])])),"dpp-policies":(0,n.w5)((()=>[(0,n.Wm)(W,{mesh:k.rawEntity.mesh,"dpp-name":k.rawEntity.name},null,8,["mesh","dpp-name"])])),"xds-configuration":(0,n.w5)((()=>[(0,n.Wm)(U,{mesh:k.rawEntity.mesh,"dpp-name":k.rawEntity.name},null,8,["mesh","dpp-name"])])),"envoy-stats":(0,n.w5)((()=>[(0,n.Wm)(z,{mesh:k.rawEntity.mesh,"dpp-name":k.rawEntity.name},null,8,["mesh","dpp-name"])])),"envoy-clusters":(0,n.w5)((()=>[(0,n.Wm)(K,{mesh:k.rawEntity.mesh,"dpp-name":k.rawEntity.name},null,8,["mesh","dpp-name"])])),mtls:(0,n.w5)((()=>[(0,n.Wm)(S,{"is-loading":k.entityIsLoading,"is-empty":k.entityIsEmpty},{default:(0,n.w5)((()=>[k.entity.mtls?((0,n.wg)(),(0,n.iD)("ul",v,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(k.entity.mtls,((e,t)=>((0,n.wg)(),(0,n.iD)("li",{key:t},[(0,n._)("h4",null,(0,n.zw)(e.label),1),(0,n._)("p",null,(0,n.zw)(e.value),1)])))),128))])):((0,n.wg)(),(0,n.j4)(R,{key:1,appearance:"danger"},{alertMessage:(0,n.w5)((()=>[b,(0,n._)("a",{href:`https://kuma.io/docs/${_.kumaDocsVersion}/documentation/security/#certificates`,class:"external-link",target:"_blank"}," Learn About Certificates in "+(0,n.zw)(k.productName),9,f)])),_:1}))])),_:1},8,["is-loading","is-empty"])])),yaml:(0,n.w5)((()=>[(0,n.Wm)(V,{"is-loading":k.entityIsLoading,"is-empty":k.entityIsEmpty,content:k.rawEntity},null,8,["is-loading","is-empty","content"])])),warnings:(0,n.w5)((()=>[(0,n.Wm)(j,{warnings:k.warnings},null,8,["warnings"])])),_:1},8,["has-error","is-loading","tabs"])):(0,n.kq)("",!0)])),_:1})}var _=a(33907),E=a(51991),D=a(93063),C=a(46187),S=a(21180),T=a(53419),P=a(70878),x=a(65404),A=a(93480),I=a(82318),L=a(78141),W=a(34707),U=a(99997),z=a(52681),K=a(51372),R=a(45689),V=a(79197),j=a(46483),N=a(70172);const Z={class:"flex items-center justify-between"},q={key:0,class:"text-lg"},O={key:1,class:"text-lg"},B={class:"subtitle"},Y={key:0},M={key:1},$={class:"flex flex-wrap justify-end"},H={class:"policy-wrapper"},G={class:"policy-type"};function Q(e,t,a,s,i,r){const o=(0,n.up)("KIcon"),l=(0,n.up)("KPop"),p=(0,n.up)("KBadge"),u=(0,n.up)("router-link"),c=(0,n.up)("AccordionItem"),d=(0,n.up)("AccordionList"),m=(0,n.up)("KCard"),y=(0,n.up)("StatusInfo");return(0,n.wg)(),(0,n.j4)(y,{"has-error":i.hasError,"is-loading":i.isLoading,"is-empty":!i.hasItems},{default:(0,n.w5)((()=>[(0,n.Wm)(m,{"border-variant":"noBorder"},{body:(0,n.w5)((()=>[(0,n.Wm)(d,{"initially-open":[],"multiple-open":""},{default:(0,n.w5)((()=>[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(i.items,((e,t)=>((0,n.wg)(),(0,n.j4)(c,{key:t},{"accordion-header":(0,n.w5)((()=>[(0,n._)("div",Z,[(0,n._)("div",null,["dataplane"===e.type?((0,n.wg)(),(0,n.iD)("p",q," Dataplane ")):(0,n.kq)("",!0),"dataplane"!==e.type?((0,n.wg)(),(0,n.iD)("p",O,(0,n.zw)(e.service),1)):(0,n.kq)("",!0),(0,n._)("p",B,["inbound"===e.type||"outbound"===e.type?((0,n.wg)(),(0,n.iD)("span",Y,(0,n.zw)(e.type)+" "+(0,n.zw)(e.name),1)):"service"===e.type||"dataplane"===e.type?((0,n.wg)(),(0,n.iD)("span",M,(0,n.zw)(e.type),1)):(0,n.kq)("",!0),(0,n.Wm)(l,{width:"300",placement:"right",trigger:"hover"},{content:(0,n.w5)((()=>[(0,n._)("div",null,(0,n.zw)(i.POLICY_TYPE_SUBTITLE[e.type]),1)])),default:(0,n.w5)((()=>[(0,n.Wm)(o,{icon:"help",size:"16",class:"ml-1"})])),_:2},1024)])]),(0,n._)("div",$,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(e.matchedPolicies,((e,a)=>((0,n.wg)(),(0,n.j4)(p,{key:`${t}-${a}`,class:"mr-2 mb-2"},{default:(0,n.w5)((()=>[(0,n.Uk)((0,n.zw)(a),1)])),_:2},1024)))),128))])])])),"accordion-content":(0,n.w5)((()=>[(0,n._)("div",H,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(e.policyTypes,((e,a)=>((0,n.wg)(),(0,n.iD)("div",{key:`${t}-${a}`,class:"policy-item"},[(0,n._)("h4",G,(0,n.zw)(e.pluralDisplayName),1),(0,n._)("ul",null,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(e.policies,((e,s)=>((0,n.wg)(),(0,n.iD)("li",{key:`${t}-${a}-${s}`,class:"my-1","data-testid":"policy-name"},[(0,n.Wm)(u,{to:e.route},{default:(0,n.w5)((()=>[(0,n.Uk)((0,n.zw)(e.name),1)])),_:2},1032,["to"])])))),128))])])))),128))])])),_:2},1024)))),128))])),_:1})])),_:1})])),_:1},8,["has-error","is-loading","is-empty"])}var F=a(29484);const X={inbound:"Policies applied on incoming connection on address",outbound:"Policies applied on outgoing connection to the address",service:"Policies applied on outgoing connections to service",dataplane:"Policies applied on all incoming and outgoing connections to the selected data plane proxy"};var J={name:"DataplanePolicies",components:{StatusInfo:F.Z,AccordionList:V.Z,AccordionItem:j.Z},props:{mesh:{type:String,required:!0},dppName:{type:String,required:!0}},data(){return{items:[],hasItems:!1,isLoading:!0,hasError:!1,searchInput:"",POLICY_TYPE_SUBTITLE:X}},computed:{...(0,_.rn)({policiesByType:e=>e.policiesByType})},watch:{dppName(){this.fetchPolicies()}},mounted(){this.fetchPolicies()},methods:{async fetchPolicies(){this.hasError=!1,this.isLoading=!0;try{const{items:e,total:t,kind:a}=await S["default"].getDataplanePolicies({mesh:this.mesh,dppName:this.dppName});void 0!==a&&"SidecarDataplane"!==a||(this.processItems(e),this.items=e,this.hasItems=t>0)}catch(e){console.error(e),this.hasError=!0}finally{this.isLoading=!1}},processItems(e){for(const t of e){t.policyTypes={};for(const e in t.matchedPolicies){const a=this.policiesByType[e],n={pluralDisplayName:a.pluralDisplayName,policies:t.matchedPolicies[e]};for(const e of n.policies)e.route={name:a.path,query:{ns:e.name},params:{mesh:e.mesh}};t.policyTypes[e]=n}}}}},ee=a(83744);const te=(0,ee.Z)(J,[["render",Q],["__scopeId","data-v-052fc18b"]]);var ae=te,ne=a(87083),se=a(51030),ie=a(25911),re={name:"DataplanesView",components:{EnvoyStats:se.Z,EnvoyClusters:ie.Z,WarningsWidget:K.Z,EntityURLControl:A.Z,FrameSkeleton:I.Z,DataOverview:L.Z,TabsWidget:W.Z,YamlView:U.Z,LabelList:z.Z,AccordionList:V.Z,AccordionItem:j.Z,SubscriptionDetails:D.Z,SubscriptionHeader:C.Z,DataplanePolicies:ae,XdsConfiguration:ne.Z,StatusInfo:F.Z},props:{nsBackButtonRoute:{type:Object,default(){return{name:"dataplanes"}}},emptyStateMsg:{type:String,default:"There are no data plane proxies present."},dataplaneApiParams:{type:Object,default(){return{}}},tableHeaders:{type:Array,default(){return[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Mesh",key:"mesh"},{label:"Type",key:"type"},{label:"Tags",key:"tags"},{label:"Last Connected",key:"lastConnected"},{label:"Last Updated",key:"lastUpdated"},{label:"Total Updates",key:"totalUpdates"},{label:"Kuma DP version",key:"dpVersion"},{label:"Envoy version",key:"envoyVersion"},{key:"warnings",hideLabel:!0}]}},tabs:{type:Array,default(){return[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"DPP Insights"},{hash:"#dpp-policies",title:"Policies"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"},{hash:"#mtls",title:"Certificate Insights"},{hash:"#yaml",title:"YAML"},{hash:"#warnings",title:"Warnings"}]}}},data(){return{productName:R.sG,isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,warnings:[],tableDataIsEmpty:!1,tableData:{headers:[],data:[]},subscriptionsReversed:[],entity:{},rawEntity:{},pageSize:R.NR,next:null,shownTLSTab:!1,rawData:null}},computed:{...(0,_.Se)({environment:"config/getEnvironment",queryNamespace:"getItemQueryNamespace",multicluster:"config/getMulticlusterStatus"}),dataplaneWizardRoute(){return"universal"===this.environment?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}},kumaDocsVersion(){const e=this.$store.getters.getKumaDocsVersion;return null!==e?e:"latest"},entityName(){return this.entity?.basicData?.name||""},entityMesh(){return this.entity?.basicData?.mesh||""}},watch:{$route(){this.loadData()}},beforeMount(){this.loadData()},methods:{onCreateClick(){E.fy.logger.info(P.T.CREATE_DATA_PLANE_PROXY_CLICKED)},buildEntity(e,t,a,n){const s=a.mTLS?(0,x.Xj)(a.mTLS):null;return{basicData:e,tags:t,mtls:s,versions:n}},init(){this.loadData()},getEmptyState(){return{title:"No Data",message:this.emptyStateMsg}},filterTabs(){return this.warnings.length?this.tabs:this.tabs.filter((e=>"#warnings"!==e.hash))},buildTableData(){return{...this.tableData,headers:this.tableHeaders}},compatibilityKind(e){return(0,x.JD)(e)},tableAction(e){const t=e;this.getEntity(t)},async parseData(e){const{dataplane:t={},dataplaneInsight:a={}}=e,{name:n="",mesh:s=""}=e,{subscriptions:i=[]}=a,r=(0,x.wY)(t),{status:o}=(0,x.lR)(t,a),{totalUpdates:l,totalRejectedUpdates:p,dpVersion:u,envoyVersion:c,selectedTime:d,selectedUpdateTime:m,version:y}=i.reduce(((e,t)=>{const{status:a={},connectTime:n,version:s={}}=t,{total:i={},lastUpdateTime:r}=a,{responsesSent:o="0",responsesRejected:l="0"}=i,{kumaDp:p={},envoy:u={}}=s,{version:c}=p,{version:d}=u;let{selectedTime:m,selectedUpdateTime:y}=e;const h=Date.parse(n),g=Date.parse(r);return h&&(!m||h>m)&&(m=h),g&&(!y||g>y)&&(y=g),{totalUpdates:e.totalUpdates+parseInt(o,10),totalRejectedUpdates:e.totalRejectedUpdates+parseInt(l,10),dpVersion:c||e.dpVersion,envoyVersion:d||e.envoyVersion,selectedTime:m,selectedUpdateTime:y,version:s||e.version}}),{totalUpdates:0,totalRejectedUpdates:0,dpVersion:"-",envoyVersion:"-",selectedTime:NaN,selectedUpdateTime:NaN,version:{}}),h={name:n,mesh:s,tags:r,status:o,totalUpdates:l,totalRejectedUpdates:p,dpVersion:u,envoyVersion:c,withWarnings:!1,unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:m?(0,T.tV)(new Date(m).toUTCString()):"never",lastConnected:d?(0,T.tV)(new Date(d).toUTCString()):"never",type:(0,x.c1)(t)},{kind:g}=this.compatibilityKind(y);switch(g){case x.Bd:h.unsupportedEnvoyVersion=!0,h.withWarnings=!0;break;case x.ZM:h.unsupportedKumaDPVersion=!0,h.withWarnings=!0;break}if(this.multicluster){const{compatible:e}=await(0,x.nF)(r,u);e||(h.withWarnings=!0,h.kumaDpAndKumaCpMismatch=!0)}return h},async loadData(e="0"){this.isLoading=!0;const t=this.$route.params.mesh||null,a=this.$route.query.ns||null;try{const{data:n,next:s}=await(0,N.W)({getSingleEntity:S["default"].getDataplaneOverviewFromMesh.bind(S["default"]),getAllEntities:S["default"].getAllDataplaneOverviews.bind(S["default"]),getAllEntitiesFromMesh:S["default"].getAllDataplaneOverviewsFromMesh.bind(S["default"]),size:this.pageSize,offset:e,mesh:t,query:a,params:{...this.dataplaneApiParams}});if(n.length){this.next=s,this.rawData=n,this.getEntity({name:n[0].name});const e=await Promise.all(n.map((e=>this.parseData(e))));this.tableData.data=e,this.tableDataIsEmpty=!1,this.isEmpty=!1}else this.tableData.data=[],this.tableDataIsEmpty=!0,this.isEmpty=!0}catch(n){this.hasError=!0,this.isEmpty=!0,console.error(n)}finally{this.isLoading=!1}},async getEntity(e){this.entityIsLoading=!0,this.entityIsEmpty=!1;const t=this.rawData.find((t=>t.name===e.name)),a=(0,x.Ng)(t);if(a){const e=["type","name","mesh"],n=(0,x.mq)(t)||{},s=(0,x.lR)(a,n),i=(0,x.wY)(a),r=(0,x.yQ)(n),o={...(0,T.wy)(a,e),status:s};this.entity=this.buildEntity(o,i,n,r),this.warnings=[];const{subscriptions:l=[]}=n;this.subscriptionsReversed=Array.from(l).reverse(),l.length&&this.setEntityWarnings(l,i),this.rawEntity=(0,T.RV)(a)}else this.entity={},this.entityIsEmpty=!0;this.entityIsLoading=!1},async setEntityWarnings(e,t){const{version:a={}}=e[e.length-1],{kumaDp:n={},envoy:s={}}=a;if(n&&s){const e=this.compatibilityKind(a),{kind:t}=e;t!==x.dG&&t!==x.O3&&this.warnings.push(e)}if(this.multicluster){const{compatible:e,payload:a}=await(0,x.nF)(t,n.version);e||this.warnings.push({kind:x.pC,payload:a})}}}};const oe=(0,ee.Z)(re,[["render",k],["__scopeId","data-v-5bfdf844"]]);var le=oe},49237:function(e,t,a){var n,s;s=a(11665),n=function(){var e;function t(){}return t.LIST_ESCAPEES=["\\","\\\\",'\\"','"',"\0","","","","","","","","\b","\t","\n","\v","\f","\r","","","","","","","","","","","","","","","","","","",(e=String.fromCharCode)(133),e(160),e(8232),e(8233)],t.LIST_ESCAPED=["\\\\",'\\"','\\"','\\"',"\\0","\\x01","\\x02","\\x03","\\x04","\\x05","\\x06","\\a","\\b","\\t","\\n","\\v","\\f","\\r","\\x0e","\\x0f","\\x10","\\x11","\\x12","\\x13","\\x14","\\x15","\\x16","\\x17","\\x18","\\x19","\\x1a","\\e","\\x1c","\\x1d","\\x1e","\\x1f","\\N","\\_","\\L","\\P"],t.MAPPING_ESCAPEES_TO_ESCAPED=function(){var e,a,n,s;for(n={},e=a=0,s=t.LIST_ESCAPEES.length;0<=s?a<s:a>s;e=0<=s?++a:--a)n[t.LIST_ESCAPEES[e]]=t.LIST_ESCAPED[e];return n}(),t.PATTERN_CHARACTERS_TO_ESCAPE=new s("[\\x00-\\x1f]|Â|Â |â¨|â©"),t.PATTERN_MAPPING_ESCAPEES=new s(t.LIST_ESCAPEES.join("|").split("\\").join("\\\\")),t.PATTERN_SINGLE_QUOTING=new s("[\\s'\":{}[\\],&*#?]|^[-?|<>=!%@`]"),t.requiresDoubleQuoting=function(e){return this.PATTERN_CHARACTERS_TO_ESCAPE.test(e)},t.escapeWithDoubleQuotes=function(e){var t;return t=this.PATTERN_MAPPING_ESCAPEES.replace(e,function(e){return function(t){return e.MAPPING_ESCAPEES_TO_ESCAPED[t]}}(this)),'"'+t+'"'},t.requiresSingleQuoting=function(e){return this.PATTERN_SINGLE_QUOTING.test(e)},t.escapeWithSingleQuotes=function(e){return"'"+e.replace(/'/g,"''")+"'"},t}(),e.exports=n},11665:function(e){var t;t=function(){function e(e,t){var a,n,s,i,r,o,l,p,u;null==t&&(t=""),s="",r=e.length,o=null,n=0,i=0;while(i<r){if(a=e.charAt(i),"\\"===a)s+=e.slice(i,+(i+1)+1||9e9),i++;else if("("===a)if(i<r-2)if(p=e.slice(i,+(i+2)+1||9e9),"(?:"===p)i+=2,s+=p;else if("(?<"===p){n++,i+=2,l="";while(i+1<r){if(u=e.charAt(i+1),">"===u){s+="(",i++,l.length>0&&(null==o&&(o={}),o[l]=n);break}l+=u,i++}}else s+=a,n++;else s+=a;else s+=a;i++}this.rawRegex=e,this.cleanedRegex=s,this.regex=new RegExp(this.cleanedRegex,"g"+t.replace("g","")),this.mapping=o}return e.prototype.regex=null,e.prototype.rawRegex=null,e.prototype.cleanedRegex=null,e.prototype.mapping=null,e.prototype.exec=function(e){var t,a,n,s;if(this.regex.lastIndex=0,a=this.regex.exec(e),null==a)return null;if(null!=this.mapping)for(n in s=this.mapping,s)t=s[n],a[n]=a[t];return a},e.prototype.test=function(e){return this.regex.lastIndex=0,this.regex.test(e)},e.prototype.replace=function(e,t){return this.regex.lastIndex=0,e.replace(this.regex,t)},e.prototype.replaceAll=function(e,t,a){var n;null==a&&(a=0),this.regex.lastIndex=0,n=0;while(this.regex.test(e)&&(0===a||n<a))this.regex.lastIndex=0,e=e.replace(this.regex,t),n++;return[e,n]},e}(),e.exports=t}}]);