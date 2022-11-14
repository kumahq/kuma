import{D as B,cl as F,cm as Z,e as M,M as q,P as K,cn as N,k as g,co as O,s as V,o as a,j as y,c as u,w as s,a as i,b as H,z as v,l as c,t as b,F as D,n as z,i as n}from"./index.08f62f25.js";import{g as R}from"./tableDataUtils.c68a3657.js";import{D as W}from"./DataOverview.4235c7a0.js";import{E as G}from"./EnvoyData.cf514968.js";import{F as P}from"./FrameSkeleton.8231b93c.js";import{_ as j}from"./LabelList.vue_vue_type_style_index_0_lang.4ef830dc.js";import{M as U}from"./MultizoneInfo.0d81f8e1.js";import{S as X,a as J}from"./SubscriptionHeader.234f64ae.js";import{T as Q}from"./TabsWidget.d4543219.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.07a221ac.js";import"./ErrorBlock.76c3bcf7.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.5c7147d0.js";import"./TagList.9cda0a7b.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.511f1454.js";import"./_commonjsHelpers.f037b798.js";const Y={name:"ZoneIngresses",components:{AccordionItem:F,AccordionList:Z,DataOverview:W,EnvoyData:G,FrameSkeleton:P,LabelList:j,MultizoneInfo:U,SubscriptionDetails:X,SubscriptionHeader:J,TabsWidget:Q,KButton:M,KCard:q},data(){return{isLoading:!0,isEmpty:!1,error:null,empty_state:{title:"No Data",message:"There are no Zone Ingresses present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Ingress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],entity:{},rawData:[],pageSize:K,next:null,subscriptionsReversed:[]}},computed:{...N({multicluster:"config/getMulticlusterStatus"})},watch:{$route(){this.isLoading=!0,this.isEmpty=!1,this.error=null,this.init()}},beforeMount(){this.init()},methods:{init(){this.multicluster&&this.loadData()},tableAction(o){const r=o;this.getEntity(r)},async loadData(o="0"){this.isLoading=!0,this.isEmpty=!1;const r=this.$route.query.ns||null;try{const{data:t,next:p}=await R({getAllEntities:g.getAllZoneIngressOverviews.bind(g),getSingleEntity:g.getZoneIngressOverview.bind(g),size:this.pageSize,offset:o,query:r});this.next=p,t.length?(this.isEmpty=!1,this.rawData=t,this.getEntity({name:t[0].name}),this.tableData.data=t.map(e=>{const{zoneIngressInsight:l={}}=e;return{...e,...O(l)}})):(this.tableData.data=[],this.isEmpty=!0)}catch(t){t instanceof Error?this.error=t:console.error(t),this.isEmpty=!0}finally{this.isLoading=!1}},getEntity(o){var e,l;const r=["type","name"],t=this.rawData.find(h=>h.name===o.name),p=(l=(e=t==null?void 0:t.zoneIngressInsight)==null?void 0:e.subscriptions)!=null?l:[];this.subscriptionsReversed=Array.from(p).reverse(),this.entity=V(t,r)}}},$={class:"zoneingresses"},ee=c("span",{class:"custom-control-icon"}," \u2190 ",-1);function te(o,r,t,p,e,l){const h=n("MultizoneInfo"),E=n("KButton"),S=n("DataOverview"),I=n("LabelList"),k=n("SubscriptionHeader"),L=n("SubscriptionDetails"),w=n("AccordionItem"),A=n("AccordionList"),x=n("KCard"),_=n("EnvoyData"),T=n("TabsWidget"),C=n("FrameSkeleton");return a(),y("div",$,[o.multicluster===!1?(a(),u(h,{key:0})):(a(),u(C,{key:1},{default:s(()=>{var f;return[i(S,{"selected-entity-name":(f=e.entity)==null?void 0:f.name,"page-size":e.pageSize,"is-loading":e.isLoading,error:e.error,"empty-state":e.empty_state,"table-data":e.tableData,"table-data-is-empty":e.isEmpty,next:e.next,onTableAction:l.tableAction,onLoadData:r[0]||(r[0]=m=>l.loadData(m))},{additionalControls:s(()=>[o.$route.query.ns?(a(),u(E,{key:0,class:"back-button",appearance:"primary",to:{name:"zoneingresses"}},{default:s(()=>[ee,H(" View All ")]),_:1})):v("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","empty-state","table-data","table-data-is-empty","next","onTableAction"]),e.isEmpty===!1?(a(),u(T,{key:0,"has-error":e.error!==null,"is-loading":e.isLoading,tabs:e.tabs,"initial-tab-override":"overview"},{tabHeader:s(()=>[c("div",null,[c("h3",null," Zone Ingress: "+b(e.entity.name),1)])]),overview:s(()=>[i(I,null,{default:s(()=>[c("div",null,[c("ul",null,[(a(!0),y(D,null,z(e.entity,(m,d)=>(a(),y("li",{key:d},[c("h4",null,b(d),1),c("p",null,b(m),1)]))),128))])])]),_:1})]),insights:s(()=>[i(x,{"border-variant":"noBorder"},{body:s(()=>[i(A,{"initially-open":0},{default:s(()=>[(a(!0),y(D,null,z(e.subscriptionsReversed,(m,d)=>(a(),u(w,{key:d},{"accordion-header":s(()=>[i(k,{details:m},null,8,["details"])]),"accordion-content":s(()=>[i(L,{details:m,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),"xds-configuration":s(()=>[i(_,{"data-path":"xds","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-stats":s(()=>[i(_,{"data-path":"stats","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-clusters":s(()=>[i(_,{"data-path":"clusters","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),_:1},8,["has-error","is-loading","tabs"])):v("",!0)]}),_:1}))])}const _e=B(Y,[["render",te]]);export{_e as default};
