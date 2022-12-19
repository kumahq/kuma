import{H as N,cm as x,cn as H,S as K,A as F,O as M,P,co as q,cp as G,cq as z,k as h,x as R,cr as J,cs as j,i as l,o as c,j as m,c as y,w as i,a as g,b,B as _,l as E,t as v,F as C,n as L}from"./index.782e29ff.js";import{_ as Y}from"./CodeBlock.vue_vue_type_style_index_0_lang.7bd622f2.js";import{D as Q,p as U}from"./patchQueryParam.d8aaed4f.js";import{F as X}from"./FrameSkeleton.d9ecfe8d.js";import{_ as $}from"./LabelList.vue_vue_type_style_index_0_lang.bff9dfc9.js";import{M as ee}from"./MultizoneInfo.2e9d5113.js";import{S as te,a as se}from"./SubscriptionHeader.4985de97.js";import{T as ne}from"./TabsWidget.d5dfcc63.js";import{W as ie}from"./WarningsWidget.07c46008.js";import"./_commonjsHelpers.f037b798.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang.fe3464f0.js";import"./ErrorBlock.e7a95361.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.10b3a197.js";import"./StatusBadge.65e74dea.js";import"./TagList.8b6a3091.js";const oe={name:"ZonesView",components:{AccordionItem:x,AccordionList:H,CodeBlock:Y,DataOverview:Q,FrameSkeleton:X,LabelList:$,MultizoneInfo:ee,SubscriptionDetails:te,SubscriptionHeader:se,TabsWidget:ne,WarningsWidget:ie,KBadge:K,KButton:F,KCard:M},props:{offset:{type:Number,required:!1,default:0}},data(){return{isLoading:!0,isEmpty:!1,error:null,entityIsLoading:!0,entityIsEmpty:!1,entityHasError:!1,tableDataIsEmpty:!1,empty_state:{title:"No Data",message:"There are no Zones present."},tableData:{headers:[{key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Storage type",key:"storeType"},{label:"Ingress",key:"hasIngress"},{label:"Egress",key:"hasEgress"},{key:"warnings",hideLabel:!0}],data:[]},tabs:[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],entity:{},pageSize:P,next:null,warnings:[],subscriptionsReversed:[],codeOutput:null,zonesWithIngress:new Set,pageOffset:this.offset}},computed:{...q({multicluster:"config/getMulticlusterStatus",globalCpVersion:"config/getVersion"})},watch:{$route(){this.isLoading=!0,this.isEmpty=!1,this.error=null,this.entityIsLoading=!0,this.entityIsEmpty=!1,this.entityHasError=!1,this.tableDataIsEmpty=!1,this.init(0)}},beforeMount(){this.init(this.offset)},methods:{init(t){this.multicluster&&this.loadData(t)},filterTabs(){return this.warnings.length?this.tabs:this.tabs.filter(t=>t.hash!=="#warnings")},tableAction(t){const s=t;this.getEntity(s)},parseData(t){const{zoneInsight:s={},name:o}=t;let n="-",e="",a=!0;s.subscriptions&&s.subscriptions.length&&s.subscriptions.forEach(u=>{if(u.version&&u.version.kumaCp){n=u.version.kumaCp.version;const{kumaCpGlobalCompatible:d=!0}=u.version.kumaCp;a=d,u.config&&(e=JSON.parse(u.config).store.type)}});const r=G(s);return{...t,status:r,zoneCpVersion:n,storeType:e,hasIngress:this.zonesWithIngress.has(o)?"Yes":"No",hasEgress:this.zonesWithEgress.has(o)?"Yes":"No",withWarnings:!a}},calculateZonesWithIngress(t){const s=new Set;t.forEach(({zoneIngress:{zone:o}})=>{s.add(o)}),this.zonesWithIngress=s},calculateZonesWithEgress(t){const s=new Set;t.forEach(({zoneEgress:{zone:o}})=>{s.add(o)}),this.zonesWithEgress=s},async loadData(t){this.pageOffset=t,U("offset",t>0?t:null),this.isLoading=!0,this.isEmpty=!1;const s=this.$route.query.ns||null,o=this.pageSize;try{const[{data:n,next:e},{items:a},{items:r}]=await Promise.all([this.getZoneOverviews(s,o,t),z(h.getAllZoneIngressOverviews.bind(h)),z(h.getAllZoneEgressOverviews.bind(h))]);this.next=e,n.length?(this.calculateZonesWithIngress(a),this.calculateZonesWithEgress(r),this.tableData.data=n.map(this.parseData),this.tableDataIsEmpty=!1,this.isEmpty=!1,this.getEntity({name:n[0].name})):(this.tableData.data=[],this.tableDataIsEmpty=!0,this.isEmpty=!0,this.entityIsEmpty=!0)}catch(n){n instanceof Error?this.error=n:console.error(n),this.isEmpty=!0}finally{this.isLoading=!1}},async getEntity(t){var n,e;this.entityIsLoading=!0,this.entityIsEmpty=!0;const s=["type","name"],o=setTimeout(()=>{this.entityIsEmpty=!0,this.entityIsLoading=!1},"500");if(t){this.entityIsEmpty=!1,this.warnings=[];try{const a=await h.getZoneOverview({name:t.name}),r=(e=(n=a.zoneInsight)==null?void 0:n.subscriptions)!=null?e:[];if(this.entity={...R(a,s),"Authentication Type":J(a)},this.subscriptionsReversed=Array.from(r).reverse(),r.length){const{version:u={}}=r[r.length-1],{kumaCp:d={}}=u,w=d.version||"-",{kumaCpGlobalCompatible:I=!0}=d;I||this.warnings.push({kind:j,payload:{zoneCpVersion:w,globalCpVersion:this.globalCpVersion}}),r[r.length-1].config&&(this.codeOutput=JSON.stringify(JSON.parse(r[r.length-1].config),null,2))}}catch(a){console.error(a),this.entity={},this.entityHasError=!0,this.entityIsEmpty=!0}finally{clearTimeout(o)}}this.entityIsLoading=!1},async getZoneOverviews(t,s,o){if(t)return{data:[await h.getZoneOverview({name:t},{size:s,offset:o})],next:null};{const{items:n,next:e}=await h.getAllZoneOverviews({size:s,offset:o});return{data:n!=null?n:[],next:e}}}}},ae={class:"zones"},re={key:0},le={key:1},ce={key:2};function pe(t,s,o,n,e,a){const r=l("MultizoneInfo"),u=l("KButton"),d=l("DataOverview"),w=l("KBadge"),I=l("LabelList"),O=l("SubscriptionHeader"),D=l("SubscriptionDetails"),A=l("AccordionItem"),W=l("AccordionList"),k=l("KCard"),Z=l("CodeBlock"),B=l("WarningsWidget"),T=l("TabsWidget"),V=l("FrameSkeleton");return c(),m("div",ae,[t.multicluster===!1?(c(),y(r,{key:0})):(c(),y(V,{key:1},{default:i(()=>{var S;return[g(d,{"selected-entity-name":(S=e.entity)==null?void 0:S.name,"page-size":e.pageSize,"is-loading":e.isLoading,error:e.error,"empty-state":e.empty_state,"table-data":e.tableData,"table-data-is-empty":e.tableDataIsEmpty,"show-warnings":e.tableData.data.some(p=>p.withWarnings),next:e.next,"page-offset":e.pageOffset,onTableAction:a.tableAction,onLoadData:s[0]||(s[0]=p=>a.loadData(p))},{additionalControls:i(()=>[t.$route.query.ns?(c(),y(u,{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"zones"}},{default:i(()=>[b(`
            View all
          `)]),_:1})):_("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","empty-state","table-data","table-data-is-empty","show-warnings","next","page-offset","onTableAction"]),b(),e.isEmpty===!1?(c(),y(T,{key:0,"has-error":e.error,"is-loading":e.isLoading,tabs:a.filterTabs(),"initial-tab-override":"overview"},{tabHeader:i(()=>[E("div",null,[E("h1",null,"Zone: "+v(e.entity.name),1)])]),overview:i(()=>[g(I,{"has-error":e.entityHasError,"is-loading":e.entityIsLoading,"is-empty":e.entityIsEmpty},{default:i(()=>[E("div",null,[E("ul",null,[(c(!0),m(C,null,L(e.entity,(p,f)=>(c(),m("li",{key:f},[p?(c(),m("h4",re,v(f),1)):_("",!0),b(),f==="status"?(c(),m("p",le,[g(w,{appearance:p==="Offline"?"danger":"success"},{default:i(()=>[b(v(p),1)]),_:2},1032,["appearance"])])):(c(),m("p",ce,v(p),1))]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])]),insights:i(()=>[g(k,{"border-variant":"noBorder"},{body:i(()=>[g(W,{"initially-open":0},{default:i(()=>[(c(!0),m(C,null,L(e.subscriptionsReversed,(p,f)=>(c(),y(A,{key:f},{"accordion-header":i(()=>[g(O,{details:p},null,8,["details"])]),"accordion-content":i(()=>[g(D,{details:p},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),config:i(()=>[e.codeOutput?(c(),y(k,{key:0,"border-variant":"noBorder"},{body:i(()=>[g(Z,{id:"code-block-zone-config",language:"json",code:e.codeOutput,"is-searchable":"","query-key":"zone-config"},null,8,["code"])]),_:1})):_("",!0)]),warnings:i(()=>[g(B,{warnings:e.warnings},null,8,["warnings"])]),_:1},8,["has-error","is-loading","tabs"])):_("",!0)]}),_:1}))])}const ze=N(oe,[["render",pe]]);export{ze as default};
