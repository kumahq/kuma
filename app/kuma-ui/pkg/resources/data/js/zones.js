"use strict";(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[798],{23382:function(e,t,n){n.r(t),n.d(t,{default:function(){return K}});var a=n(70821);const s={class:"zones"},i=(0,a._)("span",{class:"custom-control-icon"}," ← ",-1),o=(0,a.Uk)(" View All "),r={key:0},l={key:1},u={key:2},p=(0,a.Uk)(" Copy config to clipboard "),d=(0,a._)("div",null,[(0,a._)("p",null,"Config copied to clipboard!")],-1);function c(e,t,n,c,g,m){const y=(0,a.up)("MultizoneInfo"),h=(0,a.up)("KButton"),w=(0,a.up)("DataOverview"),b=(0,a.up)("EntityURLControl"),k=(0,a.up)("KBadge"),f=(0,a.up)("LabelList"),v=(0,a.up)("SubscriptionHeader"),E=(0,a.up)("SubscriptionDetails"),D=(0,a.up)("AccordionItem"),z=(0,a.up)("AccordionList"),W=(0,a.up)("KCard"),Z=(0,a.up)("CodeBlock"),_=(0,a.up)("KPop"),C=(0,a.up)("KClipboardProvider"),I=(0,a.up)("WarningsWidget"),L=(0,a.up)("TabsWidget"),S=(0,a.up)("FrameSkeleton");return(0,a.wg)(),(0,a.iD)("div",s,[!1===e.multicluster?((0,a.wg)(),(0,a.j4)(y,{key:0})):((0,a.wg)(),(0,a.j4)(S,{key:1},{default:(0,a.w5)((()=>[(0,a.Wm)(w,{"page-size":g.pageSize,"has-error":g.hasError,"is-loading":g.isLoading,"empty-state":g.empty_state,"table-data":g.tableData,"table-data-is-empty":g.tableDataIsEmpty,"show-warnings":g.tableData.data.some((e=>e.withWarnings)),next:g.next,onTableAction:m.tableAction,onLoadData:t[0]||(t[0]=e=>m.loadData(e))},{additionalControls:(0,a.w5)((()=>[e.$route.query.ns?((0,a.wg)(),(0,a.j4)(h,{key:0,class:"back-button",appearance:"primary",size:"small",to:{name:"zones"}},{default:(0,a.w5)((()=>[i,o])),_:1})):(0,a.kq)("",!0)])),_:1},8,["page-size","has-error","is-loading","empty-state","table-data","table-data-is-empty","show-warnings","next","onTableAction"]),!1===g.isEmpty?((0,a.wg)(),(0,a.j4)(L,{key:0,"has-error":g.hasError,"is-loading":g.isLoading,tabs:m.filterTabs(),"initial-tab-override":"overview"},{tabHeader:(0,a.w5)((()=>[(0,a._)("div",null,[(0,a._)("h3",null," Zone: "+(0,a.zw)(g.entity.name),1)]),(0,a._)("div",null,[(0,a.Wm)(b,{name:g.entity.name},null,8,["name"])])])),overview:(0,a.w5)((()=>[(0,a.Wm)(f,{"has-error":g.entityHasError,"is-loading":g.entityIsLoading,"is-empty":g.entityIsEmpty},{default:(0,a.w5)((()=>[(0,a._)("div",null,[(0,a._)("ul",null,[((0,a.wg)(!0),(0,a.iD)(a.HY,null,(0,a.Ko)(g.entity,((e,t)=>((0,a.wg)(),(0,a.iD)("li",{key:t},[e?((0,a.wg)(),(0,a.iD)("h4",r,(0,a.zw)(t),1)):(0,a.kq)("",!0),"status"===t?((0,a.wg)(),(0,a.iD)("p",l,[(0,a.Wm)(k,{appearance:"Offline"===e?"danger":"success"},{default:(0,a.w5)((()=>[(0,a.Uk)((0,a.zw)(e),1)])),_:2},1032,["appearance"])])):((0,a.wg)(),(0,a.iD)("p",u,(0,a.zw)(e),1))])))),128))])])])),_:1},8,["has-error","is-loading","is-empty"])])),insights:(0,a.w5)((()=>[(0,a.Wm)(W,{"border-variant":"noBorder"},{body:(0,a.w5)((()=>[(0,a.Wm)(z,{"initially-open":0},{default:(0,a.w5)((()=>[((0,a.wg)(!0),(0,a.iD)(a.HY,null,(0,a.Ko)(g.subscriptionsReversed,((e,t)=>((0,a.wg)(),(0,a.j4)(D,{key:t},{"accordion-header":(0,a.w5)((()=>[(0,a.Wm)(v,{details:e},null,8,["details"])])),"accordion-content":(0,a.w5)((()=>[(0,a.Wm)(E,{details:e},null,8,["details"])])),_:2},1024)))),128))])),_:1})])),_:1})])),config:(0,a.w5)((()=>[g.codeOutput?((0,a.wg)(),(0,a.j4)(W,{key:0,"border-variant":"noBorder"},{body:(0,a.w5)((()=>[(0,a.Wm)(Z,{language:"json",code:g.codeOutput},null,8,["code"])])),actions:(0,a.w5)((()=>[g.codeOutput?((0,a.wg)(),(0,a.j4)(C,{key:0},{default:(0,a.w5)((({copyToClipboard:e})=>[(0,a.Wm)(_,{placement:"bottom"},{content:(0,a.w5)((()=>[d])),default:(0,a.w5)((()=>[(0,a.Wm)(h,{appearance:"primary",onClick:()=>{e(g.codeOutput)}},{default:(0,a.w5)((()=>[p])),_:2},1032,["onClick"])])),_:2},1024)])),_:1})):(0,a.kq)("",!0)])),_:1})):(0,a.kq)("",!0)])),warnings:(0,a.w5)((()=>[(0,a.Wm)(I,{warnings:g.warnings},null,8,["warnings"])])),_:1},8,["has-error","is-loading","tabs"])):(0,a.kq)("",!0)])),_:1}))])}var g=n(33907),m=n(27361),y=n.n(m),h=n(93063),w=n(46187),b=n(55602),k=n(21743),f=n(53419),v=n(21180),E=n(70172),D=n(65404),z=n(82318),W=n(78141),Z=n(34707),_=n(79197),C=n(46483),I=n(93480),L=n(52681),S=n(51372),O=n(45689),U={name:"ZonesView",components:{AccordionList:_.Z,AccordionItem:C.Z,FrameSkeleton:z.Z,DataOverview:W.Z,TabsWidget:Z.Z,LabelList:L.Z,WarningsWidget:S.Z,CodeBlock:k.Z,SubscriptionDetails:h.Z,SubscriptionHeader:w.Z,MultizoneInfo:b.Z,EntityURLControl:I.Z},data(){return{isLoading:!0,isEmpty:!1,hasError:!1,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Zones present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Backend",key:"backend"},{label:"Ingress",key:"hasIngress"},{label:"Egress",key:"hasEgress"},{key:"warnings",hideLabel:!0}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],entity:{},pageSize:O.NR,next:null,warnings:[],subscriptionsReversed:[],codeOutput:null,zonesWithIngress:new Set}},computed:{...(0,g.Se)({multicluster:"config/getMulticlusterStatus",globalCpVersion:"config/getVersion"})},watch:{$route(){this.init()}},beforeMount(){this.init()},methods:{init(){this.multicluster&&this.loadData()},filterTabs(){return this.warnings.length?this.tabs:this.tabs.filter((e=>"#warnings"!==e.hash))},tableAction(e){const t=e;this.getEntity(t)},parseData(e){const{zoneInsight:t={},name:n}=e;let a="-",s="",i=!0;return t.subscriptions&&t.subscriptions.length&&t.subscriptions.forEach(((e,t)=>{if(e.version&&e.version.kumaCp){a=e.version.kumaCp.version;const{kumaCpGlobalCompatible:t=!0}=e.version.kumaCp;i=t,e.config&&(s=JSON.parse(e.config).store.type)}})),{...e,status:(0,D._I)(t).status,zoneCpVersion:a,backend:s,hasIngress:this.zonesWithIngress.has(n)?"Yes":"No",hasEgress:this.zonesWithEgress.has(n)?"Yes":"No",withWarnings:!i}},calculateZonesWithIngress(e){const t=new Set;e.forEach((({zoneIngress:{zone:e}})=>{t.add(e)})),this.zonesWithIngress=t},calculateZonesWithEgress(e){const t=new Set;e.forEach((({zoneEgress:{zone:e}})=>{t.add(e)})),this.zonesWithEgress=t},async loadData(e="0"){this.isLoading=!0,this.isEmpty=!1;const t=this.$route.query.ns||null;try{const[{data:n,next:a},{items:s},{items:i}]=await Promise.all([(0,E.W)({getSingleEntity:v["default"].getZoneOverview.bind(v["default"]),getAllEntities:v["default"].getAllZoneOverviews.bind(v["default"]),size:this.pageSize,offset:e,query:t}),(0,f.A1)({callEndpoint:v["default"].getAllZoneIngressOverviews.bind(v["default"])}),(0,f.A1)({callEndpoint:v["default"].getAllZoneEgressOverviews.bind(v["default"])})]);this.next=a,n.length?(this.calculateZonesWithIngress(s),this.calculateZonesWithEgress(i),this.tableData.data=n.map(this.parseData),this.tableDataIsEmpty=!1,this.isEmpty=!1,this.getEntity({name:n[0].name})):(this.tableData.data=[],this.tableDataIsEmpty=!0,this.isEmpty=!0,this.entityIsEmpty=!0)}catch(n){this.hasError=!0,this.isEmpty=!0,console.error(n)}finally{this.isLoading=!1}},async getEntity(e){this.entityIsLoading=!0,this.entityIsEmpty=!0;const t=["type","name"],n=setTimeout((()=>{this.entityIsEmpty=!0,this.entityIsLoading=!1}),"500");if(e){this.entityIsEmpty=!1,this.warnings=[];try{const a=await v["default"].getZoneOverview({name:e.name}),s=y()(a,"zoneInsight.subscriptions",[]);if(this.entity={...(0,f.wy)(a,t),"Authentication Type":(0,f.be)(a)},this.subscriptionsReversed=Array.from(s).reverse(),s.length){const{version:e={}}=s[s.length-1],{kumaCp:t={}}=e,n=t.version||"-",{kumaCpGlobalCompatible:a=!0}=t;a||this.warnings.push({kind:D.s9,payload:{zoneCpVersion:n,globalCpVersion:this.globalCpVersion}}),s[s.length-1].config&&(this.codeOutput=JSON.stringify(JSON.parse(s[s.length-1].config),null,2))}}catch(a){console.error(a),this.entity={},this.entityHasError=!0,this.entityIsEmpty=!0}finally{clearTimeout(n)}}this.entityIsLoading=!1}}},A=n(83744);const V=(0,A.Z)(U,[["render",c]]);var K=V},55602:function(e,t,n){n.d(t,{Z:function(){return c}});var a=n(70821);const s=(0,a._)("p",null,[(0,a.Uk)(" To access this page, you must be running in "),(0,a._)("strong",null,"Multi-Zone"),(0,a.Uk)(" mode. ")],-1),i=(0,a.Uk)(" Learn More ");function o(e,t,n,o,r,l){const u=(0,a.up)("KIcon"),p=(0,a.up)("KButton"),d=(0,a.up)("KEmptyState");return(0,a.wg)(),(0,a.j4)(d,null,{title:(0,a.w5)((()=>[(0,a.Wm)(u,{class:"kong-icon--centered",icon:"dangerCircle",size:"42"}),(0,a.Uk)(" "+(0,a.zw)(r.productName)+" is running in Standalone mode. ",1)])),message:(0,a.w5)((()=>[s])),cta:(0,a.w5)((()=>[(0,a.Wm)(p,{to:`https://kuma.io/docs/${e.kumaDocsVersion}/documentation/deployments/`,target:"_blank",appearance:"primary"},{default:(0,a.w5)((()=>[i])),_:1},8,["to"])])),_:1})}var r=n(33907),l=n(45689),u={name:"MultizoneInfo",data(){return{productName:l.sG}},computed:{...(0,r.Se)({kumaDocsVersion:"config/getKumaDocsVersion"})}},p=n(83744);const d=(0,p.Z)(u,[["render",o]]);var c=d},51372:function(e,t,n){n.d(t,{Z:function(){return B}});var a=n(70821);function s(e,t,n,s,i,o){const r=(0,a.up)("KAlert"),l=(0,a.up)("KCard");return(0,a.wg)(),(0,a.j4)(l,{"border-variant":"noBorder"},{body:(0,a.w5)((()=>[(0,a._)("ul",null,[((0,a.wg)(!0),(0,a.iD)(a.HY,null,(0,a.Ko)(n.warnings,(({kind:e,payload:t,index:n})=>((0,a.wg)(),(0,a.iD)("li",{key:`${e}/${n}`,class:"mb-1"},[(0,a.Wm)(r,{appearance:"warning"},{alertMessage:(0,a.w5)((()=>[((0,a.wg)(),(0,a.j4)((0,a.LL)(o.getWarningComponent(e)),{payload:t},null,8,["payload"]))])),_:2},1024)])))),128))])])),_:1})}function i(e,t,n,s,i,o){return(0,a.wg)(),(0,a.iD)("span",null,(0,a.zw)(n.payload),1)}var o={name:"WarningDefault",props:{payload:{type:[String,Object],required:!0}}},r=n(83744);const l=(0,r.Z)(o,[["render",i]]);var u=l;const p=(0,a.Uk)(" Envoy ("),d=(0,a.Uk)(") is unsupported by the current version of Kuma DP ("),c=(0,a.Uk)(") [Requirements: "),g=(0,a.Uk)("] ");function m(e,t,n,s,i,o){return(0,a.wg)(),(0,a.iD)("span",null,[p,(0,a._)("strong",null,(0,a.zw)(n.payload.envoy),1),d,(0,a._)("strong",null,(0,a.zw)(n.payload.kumaDp),1),c,(0,a._)("strong",null,(0,a.zw)(n.payload.requirements),1),g])}var y={name:"WarningEnvoyIncompatible",props:{payload:{type:Object,required:!0}}};const h=(0,r.Z)(y,[["render",m]]);var w=h;const b=(0,a.Uk)(" There is mismatch between versions of Kuma DP ("),k=(0,a.Uk)(") and the Zone CP ("),f=(0,a.Uk)(") ");function v(e,t,n,s,i,o){return(0,a.wg)(),(0,a.iD)("span",null,[b,(0,a._)("strong",null,(0,a.zw)(n.payload.kumaDpVersion),1),k,(0,a._)("strong",null,(0,a.zw)(n.payload.zoneVersion),1),f])}var E={name:"WarningZoneAndKumaDPVersionsIncompatible",props:{payload:{type:Object,required:!0}}};const D=(0,r.Z)(E,[["render",v]]);var z=D;const W=(0,a.Uk)(" Unsupported version of Kuma DP ("),Z=(0,a.Uk)(") ");function _(e,t,n,s,i,o){return(0,a.wg)(),(0,a.iD)("span",null,[W,(0,a._)("strong",null,(0,a.zw)(n.payload.kumaDpVersion),1),Z])}var C={name:"WarningUnsupportedKumaDPVersion",props:{payload:{type:Object,required:!0}}};const I=(0,r.Z)(C,[["render",_]]);var L=I;const S=(0,a.Uk)(" There is mismatch between versions of Zone CP ("),O=(0,a.Uk)(") and the Global CP ("),U=(0,a.Uk)(") ");function A(e,t,n,s,i,o){return(0,a.wg)(),(0,a.iD)("span",null,[S,(0,a._)("strong",null,(0,a.zw)(n.payload.zoneCpVersion),1),O,(0,a._)("strong",null,(0,a.zw)(n.payload.globalCpVersion),1),U])}var V={name:"WarningZoneAndGlobalCPSVersionsIncompatible",props:{payload:{type:Object,required:!0}}};const K=(0,r.Z)(V,[["render",A]]);var j=K,q=n(65404),T={name:"WarningsWidget",props:{warnings:{type:Array,required:!0}},methods:{getWarningComponent(e=""){switch(e){case q.Bd:return w;case q.ZM:return L;case q.pC:return z;case q.s9:return j;default:return u}}}};const P=(0,r.Z)(T,[["render",s]]);var B=P}}]);