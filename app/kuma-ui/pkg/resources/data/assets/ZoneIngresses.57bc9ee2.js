import{C as Z,cm as B,cn as q,P as M,co as N,k as y,cp as O,s as R,o as i,j as g,c as u,w as s,a as o,b as V,y as v,l,t as b,F as D,n as E,i as n}from"./index.438e3d4b.js";import{g as H}from"./tableDataUtils.bd4c47df.js";import{D as K}from"./DataOverview.d3cf1e01.js";import{_ as U}from"./EntityURLControl.vue_vue_type_script_setup_true_lang.9cb692f4.js";import{E as W}from"./EnvoyData.7c097340.js";import{F as G}from"./FrameSkeleton.a9fcc7af.js";import{_ as P}from"./LabelList.vue_vue_type_style_index_0_lang.f87328a0.js";import{M as j}from"./MultizoneInfo.e4b62e77.js";import{S as X,a as J}from"./SubscriptionHeader.1e00d863.js";import{T as Q}from"./TabsWidget.d3a9fdeb.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.5a7a3b48.js";import"./ErrorBlock.aea25275.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.5397142f.js";import"./TagList.9c7242fd.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.b3ce259f.js";import"./_commonjsHelpers.f037b798.js";const Y={name:"ZoneIngresses",components:{AccordionItem:B,AccordionList:q,DataOverview:K,EntityURLControl:U,EnvoyData:W,FrameSkeleton:G,LabelList:P,MultizoneInfo:j,SubscriptionDetails:X,SubscriptionHeader:J,TabsWidget:Q},data(){return{isLoading:!0,isEmpty:!1,error:null,empty_state:{title:"No Data",message:"There are no Zone Ingresses present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Ingress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],entity:{},rawData:[],pageSize:M,next:null,subscriptionsReversed:[]}},computed:{...N({multicluster:"config/getMulticlusterStatus"})},watch:{$route(){this.isLoading=!0,this.isEmpty=!1,this.error=null,this.init()}},beforeMount(){this.init()},methods:{init(){this.multicluster&&this.loadData()},tableAction(a){const r=a;this.getEntity(r)},async loadData(a="0"){this.isLoading=!0,this.isEmpty=!1;const r=this.$route.query.ns||null;try{const{data:t,next:p}=await H({getAllEntities:y.getAllZoneIngressOverviews.bind(y),getSingleEntity:y.getZoneIngressOverview.bind(y),size:this.pageSize,offset:a,query:r});this.next=p,t.length?(this.isEmpty=!1,this.rawData=t,this.getEntity({name:t[0].name}),this.tableData.data=t.map(e=>{const{zoneIngressInsight:c={}}=e;return{...e,...O(c)}})):(this.tableData.data=[],this.isEmpty=!0)}catch(t){t instanceof Error?this.error=t:console.error(t),this.isEmpty=!0}finally{this.isLoading=!1}},getEntity(a){var e,c;const r=["type","name"],t=this.rawData.find(h=>h.name===a.name),p=(c=(e=t==null?void 0:t.zoneIngressInsight)==null?void 0:e.subscriptions)!=null?c:[];this.subscriptionsReversed=Array.from(p).reverse(),this.entity=R(t,r)}}},$={class:"zoneingresses"},ee=l("span",{class:"custom-control-icon"}," \u2190 ",-1);function te(a,r,t,p,e,c){const h=n("MultizoneInfo"),z=n("KButton"),L=n("DataOverview"),S=n("EntityURLControl"),I=n("LabelList"),k=n("SubscriptionHeader"),w=n("SubscriptionDetails"),A=n("AccordionItem"),x=n("AccordionList"),C=n("KCard"),_=n("EnvoyData"),T=n("TabsWidget"),F=n("FrameSkeleton");return i(),g("div",$,[a.multicluster===!1?(i(),u(h,{key:0})):(i(),u(F,{key:1},{default:s(()=>{var f;return[o(L,{"selected-entity-name":(f=e.entity)==null?void 0:f.name,"page-size":e.pageSize,"is-loading":e.isLoading,error:e.error,"empty-state":e.empty_state,"table-data":e.tableData,"table-data-is-empty":e.isEmpty,next:e.next,onTableAction:c.tableAction,onLoadData:r[0]||(r[0]=m=>c.loadData(m))},{additionalControls:s(()=>[a.$route.query.ns?(i(),u(z,{key:0,class:"back-button",appearance:"primary",to:{name:"zoneingresses"}},{default:s(()=>[ee,V(" View All ")]),_:1})):v("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","empty-state","table-data","table-data-is-empty","next","onTableAction"]),e.isEmpty===!1?(i(),u(T,{key:0,"has-error":e.error!==null,"is-loading":e.isLoading,tabs:e.tabs,"initial-tab-override":"overview"},{tabHeader:s(()=>[l("div",null,[l("h3",null," Zone Ingress: "+b(e.entity.name),1)]),l("div",null,[o(S,{name:e.entity.name},null,8,["name"])])]),overview:s(()=>[o(I,null,{default:s(()=>[l("div",null,[l("ul",null,[(i(!0),g(D,null,E(e.entity,(m,d)=>(i(),g("li",{key:d},[l("h4",null,b(d),1),l("p",null,b(m),1)]))),128))])])]),_:1})]),insights:s(()=>[o(C,{"border-variant":"noBorder"},{body:s(()=>[o(x,{"initially-open":0},{default:s(()=>[(i(!0),g(D,null,E(e.subscriptionsReversed,(m,d)=>(i(),u(A,{key:d},{"accordion-header":s(()=>[o(k,{details:m},null,8,["details"])]),"accordion-content":s(()=>[o(w,{details:m,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),"xds-configuration":s(()=>[o(_,{"data-path":"xds","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-stats":s(()=>[o(_,{"data-path":"stats","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-clusters":s(()=>[o(_,{"data-path":"clusters","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),_:1},8,["has-error","is-loading","tabs"])):v("",!0)]}),_:1}))])}const be=Z(Y,[["render",te]]);export{be as default};
