import{k as T,c1 as B,c2 as F,P as q,bU as K,c4 as M,c5 as V,c6 as H,c9 as P,c8 as I,cc as a,o as r,c as g,i as m,w as n,a as l,b as _,j as D,e as u,bV as b,F as z,cd as w}from"./index-08ba2993.js";import{Q as S}from"./QueryParameter-70743f73.js";import{D as R}from"./DataOverview-1eb5b106.js";import{E as W}from"./EnvoyData-c0f25ff2.js";import{F as G}from"./FrameSkeleton-fa914657.js";import{_ as Q}from"./LabelList.vue_vue_type_style_index_0_lang-0cdd88fc.js";import{M as U}from"./MultizoneInfo-a0f62bfb.js";import{S as j,a as X}from"./SubscriptionHeader-4b351f66.js";import{T as J}from"./TabsWidget-7c52524a.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-cf69250c.js";import"./ErrorBlock-21576094.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-778739a1.js";import"./StatusBadge-c118c8ba.js";import"./TagList-e8e9bfa1.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-e26b650c.js";import"./_commonjsHelpers-87174ba5.js";const Y={name:"ZoneIngresses",components:{AccordionItem:B,AccordionList:F,DataOverview:R,EnvoyData:W,FrameSkeleton:G,LabelList:Q,MultizoneInfo:U,SubscriptionDetails:j,SubscriptionHeader:X,TabsWidget:J,KButton:q,KCard:K},props:{selectedZoneIngressName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},data(){return{isLoading:!0,isEmpty:!1,error:null,empty_state:{title:"No Data",message:"There are no Zone Ingresses present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Ingress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],entity:{},rawData:[],pageSize:M,next:null,subscriptionsReversed:[],pageOffset:this.offset}},computed:{...V({multicluster:"config/getMulticlusterStatus"})},watch:{$route(){this.isLoading=!0,this.isEmpty=!1,this.error=null,this.init(0)}},beforeMount(){this.init(this.offset)},methods:{init(t){this.multicluster&&this.loadData(t)},tableAction(t){const i=t;this.getEntity(i)},async loadData(t){this.pageOffset=t,S.set("offset",t>0?t:null),this.isLoading=!0,this.isEmpty=!1;const i=this.$route.query.ns||null,o=this.pageSize;try{const{data:s,next:e}=await this.getZoneIngressOverviews(i,o,t);this.next=e,s.length?(this.isEmpty=!1,this.rawData=s,this.getEntity({name:this.selectedZoneIngressName??s[0].name}),this.tableData.data=s.map(c=>{const{zoneIngressInsight:y={}}=c,h=H(y);return{...c,status:h}})):(this.tableData.data=[],this.isEmpty=!0)}catch(s){s instanceof Error?this.error=s:console.error(s),this.isEmpty=!0}finally{this.isLoading=!1}},getEntity(t){var e;const i=["type","name"],o=this.rawData.find(c=>c.name===t.name),s=((e=o==null?void 0:o.zoneIngressInsight)==null?void 0:e.subscriptions)??[];this.subscriptionsReversed=Array.from(s).reverse(),this.entity=P(o,i),S.set("zoneIngress",this.entity.name)},async getZoneIngressOverviews(t,i,o){if(t)return{data:[await I.getZoneIngressOverview({name:t},{size:i,offset:o})],next:null};{const{items:s,next:e}=await I.getAllZoneIngressOverviews({size:i,offset:o});return{data:s??[],next:e}}}}},$={class:"zoneingresses"},ee={class:"entity-heading"};function te(t,i,o,s,e,c){const y=a("MultizoneInfo"),h=a("KButton"),L=a("DataOverview"),k=a("LabelList"),E=a("SubscriptionHeader"),A=a("SubscriptionDetails"),x=a("AccordionItem"),O=a("AccordionList"),Z=a("KCard"),f=a("EnvoyData"),C=a("TabsWidget"),N=a("FrameSkeleton");return r(),g("div",$,[t.multicluster===!1?(r(),m(y,{key:0})):(r(),m(N,{key:1},{default:n(()=>{var v;return[l(L,{"selected-entity-name":(v=e.entity)==null?void 0:v.name,"page-size":e.pageSize,"is-loading":e.isLoading,error:e.error,"empty-state":e.empty_state,"table-data":e.tableData,"table-data-is-empty":e.isEmpty,next:e.next,"page-offset":e.pageOffset,onTableAction:c.tableAction,onLoadData:c.loadData},{additionalControls:n(()=>[t.$route.query.ns?(r(),m(h,{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"zoneingresses"}},{default:n(()=>[_(`
            View all
          `)]),_:1})):D("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","empty-state","table-data","table-data-is-empty","next","page-offset","onTableAction","onLoadData"]),_(),e.isEmpty===!1?(r(),m(C,{key:0,"has-error":e.error!==null,"is-loading":e.isLoading,tabs:e.tabs},{tabHeader:n(()=>[u("h1",ee,`
            Zone Ingress: `+b(e.entity.name),1)]),overview:n(()=>[l(k,null,{default:n(()=>[u("div",null,[u("ul",null,[(r(!0),g(z,null,w(e.entity,(p,d)=>(r(),g("li",{key:d},[u("h4",null,b(d),1),_(),u("p",null,b(p),1)]))),128))])])]),_:1})]),insights:n(()=>[l(Z,{"border-variant":"noBorder"},{body:n(()=>[l(O,{"initially-open":0},{default:n(()=>[(r(!0),g(z,null,w(e.subscriptionsReversed,(p,d)=>(r(),m(x,{key:d},{"accordion-header":n(()=>[l(E,{details:p},null,8,["details"])]),"accordion-content":n(()=>[l(A,{details:p,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),"xds-configuration":n(()=>[l(f,{"data-path":"xds","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-stats":n(()=>[l(f,{"data-path":"stats","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-clusters":n(()=>[l(f,{"data-path":"clusters","zone-ingress-name":e.entity.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),_:1},8,["has-error","is-loading","tabs"])):D("",!0)]}),_:1}))])}const _e=T(Y,[["render",te]]);export{_e as default};
