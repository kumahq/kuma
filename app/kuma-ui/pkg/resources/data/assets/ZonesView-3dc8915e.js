import{d as ne,l as oe,m as re,i as le,r as n,n as ie,p as ue,q as $,s as G,t as ce,v as fe,I as ge,o as r,c as g,u as m,k as h,w as o,a as i,E as pe,b as z,x as C,P as R,f as V,y as A,F as J,z as Y,T as ve,A as de,B as me,M as H}from"./index-c8ce0213.js";import{_ as he}from"./MultizoneInfo.vue_vue_type_script_setup_true_lang-85991008.js";import{_ as ye}from"./CodeBlock.vue_vue_type_style_index_0_lang-ad640f64.js";import{D as _e}from"./DataOverview-8df7b6a3.js";import{F as be}from"./FrameSkeleton-6fbe1de7.js";import{_ as we}from"./LabelList.vue_vue_type_style_index_0_lang-8d53c57d.js";import{_ as ke,S as Ee}from"./SubscriptionHeader.vue_vue_type_script_setup_true_lang-c2317949.js";import{T as Se}from"./TabsWidget-40ac9857.js";import{_ as Ie}from"./WarningsWidget.vue_vue_type_script_setup_true_lang-2d277287.js";import{Q}from"./QueryParameter-70743f73.js";import"./EmptyBlock.vue_vue_type_script_setup_true_lang-ec77a77d.js";import"./ErrorBlock-a8c0484c.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang-171e2abf.js";import"./TagList-64dd4a55.js";import"./StatusBadge-584f994a.js";const ze={class:"zones"},Ce={class:"entity-heading"},Ae={key:0},Ne={key:1},Te={key:2},Je=ne({__name:"ZonesView",props:{selectedZoneName:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(U){const N=U,f=oe(),j={title:"No Data",message:"There are no Zones present."},D=[{hash:"#overview",title:"Overview"},{hash:"#insights",title:"Zone Insights"},{hash:"#config",title:"Config"},{hash:"#warnings",title:"Warnings"}],T=re(),Z=le(),y=n(!0),p=n(!1),b=n(null),w=n(!0),_=n(!1),k=n(!1),E=n(!1),S=n({headers:[{label:"Actions",key:"actions",hideLabel:!0},{label:"Status",key:"status"},{label:"Name",key:"name"},{label:"Zone CP Version",key:"zoneCpVersion"},{label:"Storage type",key:"storeType"},{label:"Ingress",key:"hasIngress"},{label:"Egress",key:"hasEgress"},{label:"Warnings",key:"warnings",hideLabel:!0}],data:[]}),v=n(null),L=n(null),I=n([]),B=n([]),O=n(null),x=n(N.offset),W=n(new Set),P=n(new Set);ie(()=>T.params.mesh,function(){T.name==="zones"&&(y.value=!0,p.value=!1,w.value=!0,_.value=!1,k.value=!1,E.value=!1,b.value=null,M(0))}),ue(function(){M(N.offset)});function M(a){Z.getters["config/getMulticlusterStatus"]&&q(a)}function K(){return I.value.length===0?D.filter(a=>a.hash!=="#warnings"):D}function X(a){var d;let t="-",s="",e=!0;(((d=a.zoneInsight)==null?void 0:d.subscriptions)??[]).forEach(c=>{if(c.version&&c.version.kumaCp){t=c.version.kumaCp.version;const{kumaCpGlobalCompatible:te=!0}=c.version.kumaCp;e=te,c.config&&(s=JSON.parse(c.config).store.type)}});const u=G(a.zoneInsight);return{...a,status:u,zoneCpVersion:t,storeType:s,hasIngress:W.value.has(a.name)?"Yes":"No",hasEgress:P.value.has(a.name)?"Yes":"No",withWarnings:!e}}function ee(a){const t=new Set;a.forEach(({zoneIngress:{zone:s}})=>{t.add(s)}),W.value=t}function ae(a){const t=new Set;a.forEach(({zoneEgress:{zone:s}})=>{t.add(s)}),P.value=t}async function q(a){x.value=a,Q.set("offset",a>0?a:null),y.value=!0,p.value=!1;const t=T.query.ns||null,s=R;try{const[{data:e,next:l},{items:u},{items:d}]=await Promise.all([se(t,s,a),$(f.getAllZoneIngressOverviews.bind(f)),$(f.getAllZoneEgressOverviews.bind(f))]);L.value=l,e.length?(ee(u),ae(d),S.value.data=e.map(X),E.value=!1,p.value=!1,await F({name:N.selectedZoneName??e[0].name})):(S.value.data=[],E.value=!0,p.value=!0,_.value=!0)}catch(e){e instanceof Error?b.value=e:console.error(e),p.value=!0}finally{y.value=!1}}async function F({name:a}){var t;k.value=!1,w.value=!0,_.value=!1,I.value=[];try{const s=await f.getZoneOverview({name:a}),e=((t=s.zoneInsight)==null?void 0:t.subscriptions)??[],l=G(s.zoneInsight);if(v.value={...ce(s,["type","name"]),status:l,"Authentication Type":fe(s)},Q.set("zone",a),B.value=Array.from(e).reverse(),e.length>0){const u=e[e.length-1],d=u.version.kumaCp.version||"-",{kumaCpGlobalCompatible:c=!0}=u.version.kumaCp;c||I.value.push({kind:ge,payload:{zoneCpVersion:d,globalCpVersion:Z.getters["config/getVersion"]}}),u.config&&(O.value=JSON.stringify(JSON.parse(u.config),null,2))}}catch(s){console.error(s),v.value=null,k.value=!0,_.value=!0}finally{w.value=!1}}async function se(a,t,s){if(a)return{data:[await f.getZoneOverview({name:a},{size:t,offset:s})],next:null};{const{items:e,next:l}=await f.getAllZoneOverviews({size:t,offset:s});return{data:e??[],next:l}}}return(a,t)=>(r(),g("div",ze,[m(Z).getters["config/getMulticlusterStatus"]===!1?(r(),h(he,{key:0})):(r(),h(be,{key:1},{default:o(()=>{var s;return[i(_e,{"selected-entity-name":(s=v.value)==null?void 0:s.name,"page-size":m(R),"is-loading":y.value,error:b.value,"empty-state":j,"table-data":S.value,"table-data-is-empty":E.value,"show-warnings":S.value.data.some(e=>e.withWarnings),next:L.value,"page-offset":x.value,onTableAction:F,onLoadData:q},{additionalControls:o(()=>[a.$route.query.ns?(r(),h(m(pe),{key:0,class:"back-button",appearance:"primary",icon:"arrowLeft",to:{name:"zones"}},{default:o(()=>[z(`
            View all
          `)]),_:1})):C("",!0)]),_:1},8,["selected-entity-name","page-size","is-loading","error","table-data","table-data-is-empty","show-warnings","next","page-offset"]),z(),p.value===!1&&v.value!==null?(r(),h(Se,{key:0,"has-error":b.value!==null,"is-loading":y.value,tabs:K()},{tabHeader:o(()=>[V("h1",Ce,`
            Zone: `+A(v.value.name),1)]),overview:o(()=>[i(we,{"has-error":k.value,"is-loading":w.value,"is-empty":_.value},{default:o(()=>[V("div",null,[V("ul",null,[(r(!0),g(J,null,Y(v.value,(e,l)=>(r(),g("li",{key:l},[e?(r(),g("h4",Ae,A(l),1)):C("",!0),z(),l==="status"?(r(),g("p",Ne,[i(m(ve),{appearance:e==="Offline"?"danger":"success"},{default:o(()=>[z(A(e),1)]),_:2},1032,["appearance"])])):(r(),g("p",Te,A(e),1))]))),128))])])]),_:1},8,["has-error","is-loading","is-empty"])]),insights:o(()=>[i(m(H),{"border-variant":"noBorder"},{body:o(()=>[i(de,{"initially-open":0},{default:o(()=>[(r(!0),g(J,null,Y(B.value,(e,l)=>(r(),h(me,{key:l},{"accordion-header":o(()=>[i(ke,{details:e},null,8,["details"])]),"accordion-content":o(()=>[i(Ee,{details:e},null,8,["details"])]),_:2},1024))),128))]),_:1})]),_:1})]),config:o(()=>[O.value?(r(),h(m(H),{key:0,"border-variant":"noBorder"},{body:o(()=>[i(ye,{id:"code-block-zone-config",language:"json",code:O.value,"is-searchable":"","query-key":"zone-config"},null,8,["code"])]),_:1})):C("",!0)]),warnings:o(()=>[i(Ie,{warnings:I.value},null,8,["warnings"])]),_:1},8,["has-error","is-loading","tabs"])):C("",!0)]}),_:1}))]))}});export{Je as default};
