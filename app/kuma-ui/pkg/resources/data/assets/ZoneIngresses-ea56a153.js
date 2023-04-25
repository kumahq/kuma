import{M as R}from"./kongponents.es-131ffe48.js";import{d as U,u as j,r as o,v as G,y as H,A as K,B as W,j as f,b as _,e as d,i as b,g as u,w as s,h as I,f as Z,E as N,o as i,t as q,F as L,q as O}from"./index-834ac640.js";import{_ as X}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-cea0ad2f.js";import{A as Y,a as J}from"./AccordionList-1e1780b1.js";import{D as ee}from"./DataOverview-9d4155bf.js";import{D as te,a as se}from"./DefinitionListItem-e9b95b5e.js";import{E as k}from"./EnvoyData-d0f245f3.js";import{_ as ae,S as ne}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-2d02714a.js";import{T as re}from"./TabsWidget-b8241f4a.js";import{u as oe}from"./store-bb95959d.js";import{u as ie}from"./index-0e952743.js";import{Q as B}from"./QueryParameter-70743f73.js";import"./_plugin-vue_export-helper-c27b6911.js";import"./ErrorBlock-42ddf946.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-4325e9aa.js";import"./datadogLogEvents-302eea7b.js";import"./TagList-7389a145.js";import"./StatusBadge-1d0340ff.js";import"./CodeBlock.vue_vue_type_style_index_0_lang-82e53469.js";import"./StatusInfo.vue_vue_type_script_setup_true_lang-875056ff.js";const le={class:"zoneingresses"},ue={key:1,class:"kcard-stack"},ce={class:"kcard-border"},me={class:"kcard-border"},de={class:"entity-heading"},qe=U({__name:"ZoneIngresses",props:{selectedZoneIngressName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(C){const g=C,z=ie(),M={title:"No Data",message:"There are no Zone Ingresses present."},V=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Ingress Insights"},{hash:"#xds-configuration",title:"XDS Configuration"},{hash:"#envoy-stats",title:"Stats"},{hash:"#envoy-clusters",title:"Clusters"}],p=j(),w=oe(),m=o(!0),c=o(!1),v=o(null),y=o({headers:[{label:"Status",key:"status"},{label:"Name",key:"name"}],data:[]}),l=o(null),D=o([]),S=o(null),A=o([]),x=o(g.offset);G(()=>p.params.mesh,function(){p.name==="zoneingresses"&&(m.value=!0,c.value=!1,v.value=null,h(0))}),H(function(){F(g.offset)});function F(t){w.getters["config/getMulticlusterStatus"]&&h(t)}async function h(t){x.value=t,B.set("offset",t>0?t:null),m.value=!0,c.value=!1;const a=p.query.ns||null,r=N;try{const{data:e,next:n}=await P(a,r,t);S.value=n,e.length?(c.value=!1,D.value=e,E({name:g.selectedZoneIngressName??e[0].name}),y.value.data=e.map(T=>{const{zoneIngressInsight:$={}}=T,Q=K($);return{...T,status:Q}})):(y.value.data=[],c.value=!0)}catch(e){e instanceof Error?v.value=e:console.error(e),c.value=!0}finally{m.value=!1}}function E({name:t}){var e;const a=D.value.find(n=>n.name===t),r=((e=a==null?void 0:a.zoneIngressInsight)==null?void 0:e.subscriptions)??[];A.value=Array.from(r).reverse(),l.value=W(a,["type","name"]),B.set("zoneIngress",t)}async function P(t,a,r){if(t)return{data:[await z.getZoneIngressOverview({name:t},{size:a,offset:r})],next:null};{const{items:e,next:n}=await z.getAllZoneIngressOverviews({size:a,offset:r});return{data:e??[],next:n}}}return(t,a)=>{var r;return i(),f("div",le,[_(w).getters["config/getMulticlusterStatus"]===!1?(i(),d(X,{key:0})):(i(),f("div",ue,[b("div",ce,[u(ee,{"selected-entity-name":(r=l.value)==null?void 0:r.name,"page-size":_(N),"is-loading":m.value,error:v.value,"empty-state":M,"table-data":y.value,"table-data-is-empty":c.value,next:S.value,"page-offset":x.value,onTableAction:E,onLoadData:h},{additionalControls:s(()=>[t.$route.query.ns?(i(),d(_(R),{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"zoneingresses"}},{default:s(()=>[I(`
              View all
            `)]),_:1})):Z("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","next","page-offset"])]),I(),b("div",me,[c.value===!1&&l.value!==null?(i(),d(re,{key:0,"has-error":v.value!==null,"is-loading":m.value,tabs:V},{tabHeader:s(()=>[b("h1",de,`
              Zone Ingress: `+q(l.value.name),1)]),overview:s(()=>[u(se,null,{default:s(()=>[(i(!0),f(L,null,O(l.value,(e,n)=>(i(),d(te,{key:n,term:n},{default:s(()=>[I(q(e),1)]),_:2},1032,["term"]))),128))]),_:1})]),insights:s(()=>[u(J,{"initially-open":0},{default:s(()=>[(i(!0),f(L,null,O(A.value,(e,n)=>(i(),d(Y,{key:n},{"accordion-header":s(()=>[u(ae,{details:e},null,8,["details"])]),"accordion-content":s(()=>[u(ne,{details:e,"is-discovery-subscription":""},null,8,["details"])]),_:2},1024))),128))]),_:1})]),"xds-configuration":s(()=>[u(k,{"data-path":"xds","zone-ingress-name":l.value.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-stats":s(()=>[u(k,{"data-path":"stats","zone-ingress-name":l.value.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),"envoy-clusters":s(()=>[u(k,{"data-path":"clusters","zone-ingress-name":l.value.name,"query-key":"envoy-data-zone-ingress"},null,8,["zone-ingress-name"])]),_:1},8,["has-error","is-loading"])):Z("",!0)])]))])}}});export{qe as default};
