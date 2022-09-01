"use strict";(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[639],{74608:function(e,t,a){a.d(t,{Z:function(){return h}});var n=a(70821);const s=e=>((0,n.dD)("data-v-1fa60dc9"),e=e(),(0,n.Cn)(),e),i=(0,n.Uk)(" Copy config to clipboard "),o=s((()=>(0,n._)("div",null,[(0,n._)("p",null,"Config copied to clipboard!")],-1)));function r(e,t,a,s,r,l){const d=(0,n.up)("CodeBlock"),u=(0,n.up)("KButton"),c=(0,n.up)("KPop"),p=(0,n.up)("KClipboardProvider"),g=(0,n.up)("KCard"),m=(0,n.up)("StatusInfo");return(0,n.wg)(),(0,n.j4)(m,{"has-error":null!==r.error,"is-loading":r.isLoading,error:r.error},{default:(0,n.w5)((()=>[(0,n.Wm)(g,{"border-variant":"noBorder"},{body:(0,n.w5)((()=>[(0,n.Wm)(d,{class:"panel-code-block",language:"json",code:e.content},null,8,["code"])])),actions:(0,n.w5)((()=>[e.content?((0,n.wg)(),(0,n.j4)(p,{key:0},{default:(0,n.w5)((({copyToClipboard:t})=>[(0,n.Wm)(c,{placement:"bottom"},{content:(0,n.w5)((()=>[o])),default:(0,n.w5)((()=>[(0,n.Wm)(u,{appearance:"primary",onClick:a=>t(e.content)},{default:(0,n.w5)((()=>[i])),_:2},1032,["onClick"])])),_:2},1024)])),_:1})):(0,n.kq)("",!0)])),_:1})])),_:1},8,["has-error","is-loading","error"])}var l=a(62519),d=a(21743),u=a(21180),c=a(7433),p={name:"EnvoyData",components:{CodeBlock:d.Z,KButton:l.zU,KCard:l._s,KClipboardProvider:l.tm,KPop:l.JA,StatusInfo:c.Z},props:{dataPath:{type:String,required:!0},mesh:{type:String,required:!1,default:""},dppName:{type:String,required:!1,default:""},zoneIngressName:{type:String,required:!1,default:""},zoneEgressName:{type:String,required:!1,default:""}},data(){return{isLoading:!0,error:null}},watch:{dppName(){this.fetchContent()},zoneIngressName(){this.fetchContent()},zoneEgressName(){this.fetchContent()}},mounted(){this.fetchContent()},methods:{async fetchContent(){this.error=null,this.isLoading=!0;try{let e="";""!==this.mesh&&""!==this.dppName?e=await u["default"].getDataplaneData({dataPath:this.dataPath,mesh:this.mesh,dppName:this.dppName}):""!==this.zoneIngressName?e=await u["default"].getZoneIngressData({dataPath:this.dataPath,zoneIngressName:this.zoneIngressName}):""!==this.zoneEgressName&&(e=await u["default"].getZoneEgressData({dataPath:this.dataPath,zoneEgressName:this.zoneEgressName})),this.content="string"===typeof e?e:JSON.stringify(e,null,2)}catch(e){this.error=e}finally{this.isLoading=!1}}}},g=a(83744);const m=(0,g.Z)(p,[["render",r],["__scopeId","data-v-1fa60dc9"]]);var h=m},7433:function(e,t,a){a.d(t,{Z:function(){return k}});var n=a(70821);const s=e=>((0,n.dD)("data-v-382ac7c5"),e=e(),(0,n.Cn)(),e),i={class:"status-info"},o={class:"card-icon mb-3"},r=(0,n.Uk)(" Data Loading... "),l={class:"card-icon mb-3"},d=(0,n.Uk)(" An error has occurred while trying to load this data. "),u=s((()=>(0,n._)("summary",null,"Details",-1))),c={key:0,class:"badge-list"},p={class:"card-icon mb-3"},g=(0,n.Uk)(" There is no data to display. "),m={key:3};function h(e,t,a,s,h,y){const w=(0,n.up)("KIcon"),f=(0,n.up)("KEmptyState"),v=(0,n.up)("KBadge");return(0,n.wg)(),(0,n.iD)("div",i,[a.isLoading?((0,n.wg)(),(0,n.j4)(f,{key:0,"cta-is-hidden":"","data-testid":"status-info-loading-section"},{title:(0,n.w5)((()=>[(0,n._)("div",o,[(0,n.Wm)(w,{icon:"spinner",color:"rgba(0, 0, 0, 0.1)",size:"42"})]),r])),_:1})):a.hasError?((0,n.wg)(),(0,n.iD)(n.HY,{key:1},[(0,n.Wm)(f,{"cta-is-hidden":""},(0,n.Nv)({title:(0,n.w5)((()=>[(0,n._)("div",l,[(0,n.Wm)(w,{class:"kong-icon--centered",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"42"})]),y.shouldShowApiError?((0,n.wg)(),(0,n.iD)(n.HY,{key:0},[(0,n.Uk)((0,n.zw)(a.error.message),1)],64)):((0,n.wg)(),(0,n.iD)(n.HY,{key:1},[d],64))])),_:2},[y.shouldShowApiError&&Array.isArray(a.error.causes)&&a.error.causes.length>0?{name:"message",fn:(0,n.w5)((()=>[(0,n._)("details",null,[u,(0,n._)("ul",null,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(a.error.causes,((e,t)=>((0,n.wg)(),(0,n.iD)("li",{key:t},[(0,n._)("b",null,[(0,n._)("code",null,(0,n.zw)(e.field),1)]),(0,n.Uk)(": "+(0,n.zw)(e.message),1)])))),128))])])])),key:"0"}:void 0]),1024),y.shouldShowApiError?((0,n.wg)(),(0,n.iD)("div",c,[a.error.code?((0,n.wg)(),(0,n.j4)(v,{key:0,appearance:"warning"},{default:(0,n.w5)((()=>[(0,n.Uk)((0,n.zw)(a.error.code),1)])),_:1})):(0,n.kq)("",!0),(0,n.Wm)(v,{appearance:"warning"},{default:(0,n.w5)((()=>[(0,n.Uk)((0,n.zw)(a.error.statusCode),1)])),_:1})])):(0,n.kq)("",!0)],64)):a.isEmpty?((0,n.wg)(),(0,n.j4)(f,{key:2,"cta-is-hidden":""},{title:(0,n.w5)((()=>[(0,n._)("div",p,[(0,n.Wm)(w,{class:"kong-icon--centered",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"42"})]),g])),_:1})):((0,n.wg)(),(0,n.iD)("div",m,[(0,n.WI)(e.$slots,"default",{},void 0,!0)]))])}var y=a(62519),w=a(76502),f={name:"StatusInfo",components:{KBadge:y.i1,KEmptyState:y.KB,KIcon:y.Ec},props:{isLoading:{type:Boolean,default:!1},hasError:{type:Boolean,default:!1},isEmpty:{type:Boolean,default:!1},error:{type:Object,required:!1,default:null}},computed:{shouldShowApiError(){return this.error instanceof w.ApiError}}},v=a(83744);const b=(0,v.Z)(f,[["render",h],["__scopeId","data-v-382ac7c5"]]);var k=b},96204:function(e,t,a){a.r(t),a.d(t,{default:function(){return W}});var n=a(70821);const s={class:"zoneingresses"},i=(0,n._)("span",{class:"custom-control-icon"}," ← ",-1),o=(0,n.Uk)(" View All ");function r(e,t,a,r,l,d){const u=(0,n.up)("MultizoneInfo"),c=(0,n.up)("KButton"),p=(0,n.up)("DataOverview"),g=(0,n.up)("EntityURLControl"),m=(0,n.up)("LabelList"),h=(0,n.up)("SubscriptionHeader"),y=(0,n.up)("SubscriptionDetails"),w=(0,n.up)("AccordionItem"),f=(0,n.up)("AccordionList"),v=(0,n.up)("KCard"),b=(0,n.up)("EnvoyData"),k=(0,n.up)("TabsWidget"),_=(0,n.up)("FrameSkeleton");return(0,n.wg)(),(0,n.iD)("div",s,[!1===e.multicluster?((0,n.wg)(),(0,n.j4)(u,{key:0})):((0,n.wg)(),(0,n.j4)(_,{key:1},{default:(0,n.w5)((()=>[(0,n.Wm)(p,{"page-size":l.pageSize,"has-error":l.hasError,"is-loading":l.isLoading,"empty-state":l.empty_state,"table-data":l.tableData,"table-data-is-empty":l.isEmpty,next:l.next,onTableAction:d.tableAction,onLoadData:t[0]||(t[0]=e=>d.loadData(e))},{additionalControls:(0,n.w5)((()=>[e.$route.query.ns?((0,n.wg)(),(0,n.j4)(c,{key:0,class:"back-button",appearance:"primary",size:"small",to:{name:"zoneingresses"}},{default:(0,n.w5)((()=>[i,o])),_:1})):(0,n.kq)("",!0)])),_:1},8,["page-size","has-error","is-loading","empty-state","table-data","table-data-is-empty","next","onTableAction"]),!1===l.isEmpty?((0,n.wg)(),(0,n.j4)(k,{key:0,"has-error":l.hasError,"is-loading":l.isLoading,tabs:l.tabs,"initial-tab-override":"overview"},{tabHeader:(0,n.w5)((()=>[(0,n._)("div",null,[(0,n._)("h3",null," Zone Ingress: "+(0,n.zw)(l.entity.name),1)]),(0,n._)("div",null,[(0,n.Wm)(g,{name:l.entity.name},null,8,["name"])])])),overview:(0,n.w5)((()=>[(0,n.Wm)(m,null,{default:(0,n.w5)((()=>[(0,n._)("div",null,[(0,n._)("ul",null,[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(l.entity,((e,t)=>((0,n.wg)(),(0,n.iD)("li",{key:t},[(0,n._)("h4",null,(0,n.zw)(t),1),(0,n._)("p",null,(0,n.zw)(e),1)])))),128))])])])),_:1})])),insights:(0,n.w5)((()=>[(0,n.Wm)(v,{"border-variant":"noBorder"},{body:(0,n.w5)((()=>[(0,n.Wm)(f,{"initially-open":0},{default:(0,n.w5)((()=>[((0,n.wg)(!0),(0,n.iD)(n.HY,null,(0,n.Ko)(l.subscriptionsReversed,((e,t)=>((0,n.wg)(),(0,n.j4)(w,{key:t},{"accordion-header":(0,n.w5)((()=>[(0,n.Wm)(h,{details:e},null,8,["details"])])),"accordion-content":(0,n.w5)((()=>[(0,n.Wm)(y,{details:e,"is-discovery-subscription":""},null,8,["details"])])),_:2},1024)))),128))])),_:1})])),_:1})])),"xds-configuration":(0,n.w5)((()=>[(0,n.Wm)(b,{"data-path":"xds","zone-ingress-name":l.entity.name},null,8,["zone-ingress-name"])])),"envoy-stats":(0,n.w5)((()=>[(0,n.Wm)(b,{"data-path":"stats","zone-ingress-name":l.entity.name},null,8,["zone-ingress-name"])])),"envoy-clusters":(0,n.w5)((()=>[(0,n.Wm)(b,{"data-path":"clusters","zone-ingress-name":l.entity.name},null,8,["zone-ingress-name"])])),_:1},8,["has-error","is-loading","tabs"])):(0,n.kq)("",!0)])),_:1}))])}var l=a(27361),d=a.n(l),u=a(33907),c=a(93063),p=a(46187),g=a(55602),m=a(70172),h=a(53419),y=a(21180),w=a(82318),f=a(59713),v=a(93480),b=a(18574),k=a(52681),_=a(65404),z=a(45689),D=a(79197),E=a(46483),S=a(74608),I={name:"ZoneIngresses",components:{EnvoyData:S.Z,FrameSkeleton:w.Z,DataOverview:f.Z,TabsWidget:b.Z,LabelList:k.Z,AccordionList:D.Z,AccordionItem:E.Z,SubscriptionDetails:c.Z,SubscriptionHeader:p.Z,MultizoneInfo:g.Z,EntityURLControl:v.Z},data(){return{isLoading:!0,isEmpty:!1,hasError:!1,empty_state:{title:"No Data",message:"There are no Zone Ingresses present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Ingress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],entity:{},rawData:[],pageSize:z.NR,next:null,subscriptionsReversed:[]}},computed:{...(0,u.Se)({multicluster:"config/getMulticlusterStatus"})},watch:{$route(){this.isLoading=!0,this.isEmpty=!1,this.hasError=!1,this.init()}},beforeMount(){this.init()},methods:{init(){this.multicluster&&this.loadData()},tableAction(e){const t=e;this.getEntity(t)},async loadData(e="0"){this.isLoading=!0,this.isEmpty=!1;const t=this.$route.query.ns||null;try{const{data:a,next:n}=await(0,m.W)({getAllEntities:y["default"].getAllZoneIngressOverviews.bind(y["default"]),getSingleEntity:y["default"].getZoneIngressOverview.bind(y["default"]),size:this.pageSize,offset:e,query:t});this.next=n,a.length?(this.isEmpty=!1,this.rawData=a,this.getEntity({name:a[0].name}),this.tableData.data=a.map((e=>{const{zoneIngressInsight:t={}}=e;return{...e,...(0,_._I)(t)}}))):(this.tableData.data=[],this.isEmpty=!0)}catch(a){this.hasError=!0,this.isEmpty=!0,console.error(a)}finally{this.isLoading=!1}},getEntity(e){const t=["type","name"],a=this.rawData.find((t=>t.name===e.name)),n=d()(a,"zoneIngressInsight.subscriptions",[]);this.subscriptionsReversed=Array.from(n).reverse(),this.entity=(0,h.wy)(a,t)}}},Z=a(83744);const C=(0,Z.Z)(I,[["render",r]]);var W=C},55602:function(e,t,a){a.d(t,{Z:function(){return p}});var n=a(70821);const s=(0,n._)("p",null,[(0,n.Uk)(" To access this page, you must be running in "),(0,n._)("strong",null,"Multi-Zone"),(0,n.Uk)(" mode. ")],-1),i=(0,n.Uk)(" Learn More ");function o(e,t,a,o,r,l){const d=(0,n.up)("KIcon"),u=(0,n.up)("KButton"),c=(0,n.up)("KEmptyState");return(0,n.wg)(),(0,n.j4)(c,null,{title:(0,n.w5)((()=>[(0,n.Wm)(d,{class:"kong-icon--centered",icon:"dangerCircle",size:"42"}),(0,n.Uk)(" "+(0,n.zw)(r.productName)+" is running in Standalone mode. ",1)])),message:(0,n.w5)((()=>[s])),cta:(0,n.w5)((()=>[(0,n.Wm)(u,{to:`https://kuma.io/docs/${e.kumaDocsVersion}/documentation/deployments/`,target:"_blank",appearance:"primary"},{default:(0,n.w5)((()=>[i])),_:1},8,["to"])])),_:1})}var r=a(33907),l=a(45689),d={name:"MultizoneInfo",data(){return{productName:l.sG}},computed:{...(0,r.Se)({kumaDocsVersion:"config/getKumaDocsVersion"})}},u=a(83744);const c=(0,u.Z)(d,[["render",o]]);var p=c}}]);