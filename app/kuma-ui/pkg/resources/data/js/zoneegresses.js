"use strict";(self["webpackChunkkuma_gui"]=self["webpackChunkkuma_gui"]||[]).push([[722],{14850:function(e,t,s){s.r(t),s.d(t,{default:function(){return C}});var a=s(70821);const n={class:"zoneegresses"},i=(0,a._)("span",{class:"custom-control-icon"}," ← ",-1),o=(0,a.Uk)(" View All "),r={key:0};function l(e,t,s,l,u,d){const g=(0,a.up)("KButton"),c=(0,a.up)("DataOverview"),m=(0,a.up)("EntityURLControl"),h=(0,a.up)("LabelList"),y=(0,a.up)("SubscriptionHeader"),p=(0,a.up)("SubscriptionDetails"),b=(0,a.up)("AccordionItem"),w=(0,a.up)("AccordionList"),v=(0,a.up)("KCard"),E=(0,a.up)("XdsConfiguration"),f=(0,a.up)("EnvoyStats"),D=(0,a.up)("EnvoyClusters"),k=(0,a.up)("TabsWidget"),z=(0,a.up)("FrameSkeleton");return(0,a.wg)(),(0,a.iD)("div",n,[(0,a.Wm)(z,null,{default:(0,a.w5)((()=>[(0,a.Wm)(c,{"page-size":u.pageSize,"has-error":u.hasError,"is-loading":u.isLoading,"empty-state":u.empty_state,"table-data":u.tableData,"table-data-is-empty":u.isEmpty,next:u.next,onTableAction:d.tableAction,onLoadData:t[0]||(t[0]=e=>d.loadData(e))},{additionalControls:(0,a.w5)((()=>[e.$route.query.ns?((0,a.wg)(),(0,a.j4)(g,{key:0,class:"back-button",appearance:"primary",size:"small",to:{name:"zoneegresses"}},{default:(0,a.w5)((()=>[i,o])),_:1})):(0,a.kq)("",!0)])),_:1},8,["page-size","has-error","is-loading","empty-state","table-data","table-data-is-empty","next","onTableAction"]),!1===u.isEmpty?((0,a.wg)(),(0,a.j4)(k,{key:0,"has-error":u.hasError,"is-loading":u.isLoading,tabs:u.tabs,"initial-tab-override":"overview"},{tabHeader:(0,a.w5)((()=>[(0,a._)("div",null,[(0,a._)("h3",null," Zone Egress: "+(0,a.zw)(u.entity.name),1)]),(0,a._)("div",null,[(0,a.Wm)(m,{name:u.entity.name},null,8,["name"])])])),overview:(0,a.w5)((()=>[(0,a.Wm)(h,null,{default:(0,a.w5)((()=>[(0,a._)("div",null,[(0,a._)("ul",null,[((0,a.wg)(!0),(0,a.iD)(a.HY,null,(0,a.Ko)(u.entity,((e,t)=>((0,a.wg)(),(0,a.iD)("li",{key:t},[e?((0,a.wg)(),(0,a.iD)("h4",r,(0,a.zw)(t),1)):(0,a.kq)("",!0),(0,a._)("p",null,(0,a.zw)(e),1)])))),128))])])])),_:1})])),insights:(0,a.w5)((()=>[(0,a.Wm)(v,{"border-variant":"noBorder"},{body:(0,a.w5)((()=>[(0,a.Wm)(w,{"initially-open":0},{default:(0,a.w5)((()=>[((0,a.wg)(!0),(0,a.iD)(a.HY,null,(0,a.Ko)(u.subscriptionsReversed,((e,t)=>((0,a.wg)(),(0,a.j4)(b,{key:t},{"accordion-header":(0,a.w5)((()=>[(0,a.Wm)(y,{details:e},null,8,["details"])])),"accordion-content":(0,a.w5)((()=>[(0,a.Wm)(p,{details:e,"is-discovery-subscription":""},null,8,["details"])])),_:2},1024)))),128))])),_:1})])),_:1})])),"xds-configuration":(0,a.w5)((()=>[(0,a.Wm)(E,{"zone-egress-name":u.entity.name},null,8,["zone-egress-name"])])),"envoy-stats":(0,a.w5)((()=>[(0,a.Wm)(f,{"zone-egress-name":u.entity.name},null,8,["zone-egress-name"])])),"envoy-clusters":(0,a.w5)((()=>[(0,a.Wm)(D,{"zone-egress-name":u.entity.name},null,8,["zone-egress-name"])])),_:1},8,["has-error","is-loading","tabs"])):(0,a.kq)("",!0)])),_:1})])}var u=s(27361),d=s.n(u),g=s(93063),c=s(46187),m=s(70172),h=s(53419),y=s(21180),p=s(78141),b=s(93480),w=s(82318),v=s(34707),E=s(52681),f=s(87083),D=s(65404),k=s(45689),z=s(79197),_=s(46483),Z=s(51030),L=s(25911),S={name:"ZoneEgresses",components:{EnvoyClusters:L.Z,EnvoyStats:Z.Z,FrameSkeleton:w.Z,DataOverview:p.Z,TabsWidget:v.Z,LabelList:E.Z,AccordionList:z.Z,AccordionItem:_.Z,SubscriptionDetails:g.Z,SubscriptionHeader:c.Z,EntityURLControl:b.Z,XdsConfiguration:f.Z},data(){return{isLoading:!0,isEmpty:!1,hasError:!1,empty_state:{title:"No Data",message:"There are no Zone Egresses present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Egress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],entity:{},rawData:[],pageSize:k.NR,next:null,subscriptionsReversed:[]}},watch:{$route(){this.init()}},beforeMount(){this.init()},methods:{init(){this.loadData()},tableAction(e){const t=e;this.getEntity(t)},async loadData(e="0"){this.isLoading=!0,this.isEmpty=!1;const t=this.$route.query.ns||null;try{const{data:s,next:a}=await(0,m.W)({getAllEntities:y["default"].getAllZoneEgressOverviews.bind(y["default"]),getSingleEntity:y["default"].getZoneEgressOverview.bind(y["default"]),size:this.pageSize,offset:e,query:t});this.next=a,s.length?(this.isEmpty=!1,this.rawData=s,this.getEntity({name:s[0].name}),this.tableData.data=s.map((e=>{const{zoneEgressInsight:t={}}=e;return{...e,...(0,D._I)(t)}}))):(this.tableData.data=[],this.isEmpty=!0)}catch(s){this.hasError=!0,this.isEmpty=!0,console.error(s)}finally{this.isLoading=!1}},getEntity(e){const t=["type","name"],s=this.rawData.find((t=>t.name===e.name)),a=d()(s,"zoneEgressInsight.subscriptions",[]);this.subscriptionsReversed=Array.from(a).reverse(),this.entity=(0,h.wy)(s,t)}}},W=s(83744);const A=(0,W.Z)(S,[["render",l]]);var C=A}}]);