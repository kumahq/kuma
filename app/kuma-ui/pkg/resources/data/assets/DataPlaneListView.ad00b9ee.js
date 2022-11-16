import{d as he,f as U,cz as _e,cu as me,cA as ge,cB as Ie,cC as Ke,o as i,j as v,l as e,a as O,w as S,b as G,t as g,A as Oe,u as d,z as x,F as A,n as M,K as Le,i as Ne,B as be,C as ke,D as De,p as Re,r as T,g as $e,S as ve,k as xe,c as B,m as Me,cD as ze,cE as je,cF as He,e as fe,q as Be,E as Ge,G as qe,cs as ye,cG as Ye,cH as Ze,cI as Fe,cw as Je,cx as Qe,cJ as We,cv as Xe}from"./index.09f51eb8.js";import{C as et}from"./ContentWrapper.dfc9e5ec.js";import{D as tt}from"./DataOverview.75cd0a9e.js";import{T as at}from"./TagList.c6e1c385.js";import{Y as nt}from"./YamlView.e892619a.js";import{_ as st}from"./EmptyBlock.vue_vue_type_script_setup_true_lang.264304d3.js";import"./ErrorBlock.6cb5eaea.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.f2b15057.js";import"./index.58caa11d.js";import"./CodeBlock.vue_vue_type_style_index_0_lang.f4101bd8.js";import"./_commonjsHelpers.f037b798.js";const V=a=>(be("data-v-49299a0f"),a=a(),ke(),a),lt={class:"entity-summary entity-section-list"},ot={class:"entity-title","data-testid":"data-plane-proxy-title"},it=V(()=>e("span",{class:"kutil-sr-only"},"Data plane proxy:",-1)),rt={class:"definition"},dt=V(()=>e("span",null,"Mesh:",-1)),ct={key:0},ut=V(()=>e("h4",null,"Tags",-1)),pt={key:1},mt=V(()=>e("h4",null,"Dependencies",-1)),vt={class:"mt-2 heading-with-icon"},ft=V(()=>e("h4",null,"Insights",-1)),yt={class:"entity-section-list"},ht=["data-testid"],_t=V(()=>e("span",null,"Connect time:",-1)),gt=["data-testid"],bt=V(()=>e("span",null,"Disconnect time:",-1)),kt={class:"definition"},Dt=V(()=>e("span",null,"Control plane instance ID:",-1)),wt={key:0},Tt=V(()=>e("summary",null," Responses (acknowledged / sent) ",-1)),Ct=["data-testid"],Pt=he({__name:"DataPlaneEntitySummary",props:{dataPlaneOverview:{type:Object,required:!0}},setup(a){const m=a,h={"Partially degraded":"partially_degraded",Offline:"offline",Online:"online"},F=U(()=>{const{name:p,mesh:r,dataplane:f}=m.dataPlaneOverview;return{type:"Dataplane",name:p,mesh:r,networking:f.networking}}),E=U(()=>_e(m.dataPlaneOverview.dataplane)),z=U(()=>{const p=Array.from(m.dataPlaneOverview.dataplaneInsight.subscriptions);return p.reverse(),p.map(r=>{const f=r.connectTime!==void 0?me(r.connectTime):"\u2014",l=r.disconnectTime!==void 0?me(r.disconnectTime):"\u2014",c=Object.entries(r.status).filter(([b])=>!["total","lastUpdateTime"].includes(b)).map(([b,k])=>{var N,q,Y,Z,j;const D=`${(N=k.responsesAcknowledged)!=null?N:0} / ${(q=k.responsesSent)!=null?q:0}`;return{type:b.toUpperCase(),ratio:D,responsesSent:(Y=k.responsesSent)!=null?Y:0,responsesAcknowledged:(Z=k.responsesAcknowledged)!=null?Z:0,responsesRejected:(j=k.responsesRejected)!=null?j:0}});return{subscription:r,formattedConnectDate:f,formattedDisconnectDate:l,statuses:c}})}),C=U(()=>{const{status:p}=ge(m.dataPlaneOverview.dataplane,m.dataPlaneOverview.dataplaneInsight);return Ie[h[p]]}),L=U(()=>{const p=Ke(m.dataPlaneOverview.dataplaneInsight);return p!==null?Object.entries(p).map(([r,f])=>({name:r,version:f})):[]}),I=U(()=>{const{subscriptions:p}=m.dataPlaneOverview.dataplaneInsight;if(p.length===0)return[];const r=p[p.length-1];if(!r.version)return[];const f=[],l=r.version.envoy,c=r.version.kumaDp;if(!(l.kumaDpCompatible!==void 0?l.kumaDpCompatible:!0)){const D=`Envoy ${l.version} is not supported by Kuma DP ${c.version}.`;f.push(D)}if(!(c.kumaCpCompatible!==void 0?c.kumaCpCompatible:!0)){const D=`Kuma DP ${c.version} is not supported by this Kuma control plane.`;f.push(D)}return f});return(p,r)=>{const f=Ne("router-link");return i(),v("div",lt,[e("section",null,[e("h3",ot,[it,O(f,{to:{name:"data-plane-detail-view",params:{mesh:a.dataPlaneOverview.mesh,dataPlane:a.dataPlaneOverview.name}}},{default:S(()=>[G(g(a.dataPlaneOverview.name),1)]),_:1},8,["to"]),e("div",{class:Oe(`status status--${d(C).appearance}`),"data-testid":"data-plane-status-badge"},g(d(C).title.toLowerCase()),3)]),e("div",rt,[dt,e("span",null,g(a.dataPlaneOverview.mesh),1)])]),d(E).length>0?(i(),v("section",ct,[ut,O(at,{tags:d(E)},null,8,["tags"])])):x("",!0),d(L).length>0?(i(),v("section",pt,[mt,(i(!0),v(A,null,M(d(L),(l,c)=>(i(),v("div",{key:c,class:"definition"},[e("span",null,g(l.name)+":",1),e("span",null,g(l.version),1)]))),128)),d(I).length>0?(i(),v(A,{key:0},[e("h5",vt,[G(" Warnings "),O(d(Le),{class:"ml-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"20"})]),(i(!0),v(A,null,M(d(I),(l,c)=>(i(),v("p",{key:c},g(l),1))),128))],64)):x("",!0)])):x("",!0),d(z).length>0?(i(),v(A,{key:2},[e("section",null,[ft,e("div",yt,[(i(!0),v(A,null,M(d(z),(l,c)=>(i(),v("div",{key:c},[e("div",{class:"definition","data-testid":`data-plane-connect-time-${c}`},[_t,e("span",null,g(l.formattedConnectDate),1)],8,ht),e("div",{class:"definition","data-testid":`data-plane-disconnect-time-${c}`},[bt,e("span",null,g(l.formattedDisconnectDate),1)],8,gt),e("div",kt,[Dt,e("span",null,g(l.subscription.controlPlaneInstanceId),1)]),l.statuses.length>0?(i(),v("details",wt,[Tt,(i(!0),v(A,null,M(l.statuses,(b,k)=>(i(),v("div",{key:`${c}-${k}`,class:"definition","data-testid":`data-plane-subscription-status-${c}-${k}`},[e("span",null,g(b.type)+":",1),e("span",null,g(b.ratio),1)],8,Ct))),128))])):x("",!0)]))),128))])]),e("section",null,[O(nt,{id:"code-block-data-plane-summary",content:d(F),"code-max-height":"250px"},null,8,["content"])])],64)):x("",!0)])}}});const Ut=De(Pt,[["__scopeId","data-v-49299a0f"]]),we=[{key:"status",label:"Status"},{key:"name",label:"Name"},{key:"mesh",label:"Mesh"},{key:"type",label:"Type"},{key:"service",label:"Service"},{key:"protocol",label:"Protocol"},{key:"zone",label:"Zone"},{key:"lastConnected",label:"Last Connected"},{key:"lastUpdated",label:"Last Updated"},{key:"totalUpdates",label:"Total Updates"},{key:"dpVersion",label:"Kuma DP version"},{key:"envoyVersion",label:"Envoy version"},{key:"details",label:"Details",hideLabel:!0}],Vt=["name","details"],St=we.filter(a=>!Vt.includes(a.key)).map(a=>({tableHeaderKey:a.key,label:a.label,isChecked:!1})),Te=["status","name","mesh","type","service","protocol","zone","lastUpdated","dpVersion","details"];function At(a,m=Te){return we.filter(h=>m.includes(h.key)?a?!0:h.key!=="zone":!1)}function ee(a,m){const h=new URL(window.location.href);m!=null?h.searchParams.set(a,String(m)):h.searchParams.has(a)&&h.searchParams.delete(a),window.history.replaceState({path:h.href},"",h.href)}const te=a=>(be("data-v-ff8c3ecb"),a=a(),ke(),a),Et=te(()=>e("label",{for:"data-planes-type-filter",class:"mr-2"}," Type: ",-1)),It=["value"],Kt=["for"],Ot=["id","checked","onChange"],Lt=te(()=>e("span",{class:"custom-control-icon"}," + ",-1)),Nt=te(()=>e("span",{class:"custom-control-icon"}," \u2190 ",-1)),Rt=he({__name:"DataPlaneListView",props:{name:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(a){const m=a,h=50,F=["All","Standard","Gateway (builtin)","Gateway (delegated)"],E=Re(),z=Be(),C=T(Te),L=T(!0),I=T(!1),p=T(null),r=T(!1),f=T({headers:[],data:[]}),l=T([]),c=T(null),b=T("All"),k=T(m.offset),D=T(null),N=U(()=>z.getters["config/getMulticlusterStatus"]),q=U(()=>({name:z.getters["config/getEnvironment"]==="universal"?"universal-dataplane":"kubernetes-dataplane"})),Y=U(()=>{const t=f.value.data.filter(y=>b.value==="All"?!0:y.type.toLowerCase()===b.value.toLowerCase()),s=At(N.value,C.value);return{data:t,headers:s}}),Z=U(()=>St.filter(t=>N.value?!0:t.tableHeaderKey!=="zone").map(t=>{const s=C.value.includes(t.tableHeaderKey);return{...t,isChecked:s}}));$e(()=>E.params.mesh,function(){E.name==="data-plane-list-view"&&(I.value=!1,p.value=null,r.value=!1,J(0))});const j=ve.get("dpVisibleTableHeaderKeys");Array.isArray(j)&&(C.value=j),J(m.offset);function Ce(t){t.stopPropagation()}function Pe(t,s){const y=t.target,o=C.value.findIndex(u=>u===s);y.checked&&o===-1?C.value.push(s):!y.checked&&o>-1&&C.value.splice(o,1),ve.set("dpVisibleTableHeaderKeys",Array.from(new Set(C.value)))}function Ue(){Ge.logger.info(qe.CREATE_DATA_PLANE_PROXY_CLICKED)}function Ve(){return{title:"No Data",message:"There are no data plane proxies present."}}async function Se(t){var ne,se,le,oe,ie;const s=t.mesh,y=t.name,o={name:"data-plane-detail-view",params:{mesh:s,dataPlane:y}},u={name:"mesh-detail-view",params:{mesh:s}},H=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],R=_e(t.dataplane).filter(n=>H.includes(n.label)),w=(ne=R.find(n=>n.label==="kuma.io/service"))==null?void 0:ne.value,K=(se=R.find(n=>n.label==="kuma.io/protocol"))==null?void 0:se.value,W=(le=R.find(n=>n.label==="kuma.io/zone"))==null?void 0:le.value;let ae;w!==void 0&&(ae={name:"service-insight-detail-view",params:{mesh:s,service:w}});const{status:Ae}=ge(t.dataplane,t.dataplaneInsight),Ee={totalUpdates:0,totalRejectedUpdates:0,dpVersion:null,envoyVersion:null,selectedTime:NaN,selectedUpdateTime:NaN,version:null},_=t.dataplaneInsight.subscriptions.reduce((n,P)=>{var re,de,ce,ue;if(P.connectTime){const pe=Date.parse(P.connectTime);(!n.selectedTime||pe>n.selectedTime)&&(n.selectedTime=pe)}const X=Date.parse(P.status.lastUpdateTime);return X&&(!n.selectedUpdateTime||X>n.selectedUpdateTime)&&(n.selectedUpdateTime=X),{totalUpdates:n.totalUpdates+parseInt((re=P.status.total.responsesSent)!=null?re:"0",10),totalRejectedUpdates:n.totalRejectedUpdates+parseInt((de=P.status.total.responsesRejected)!=null?de:"0",10),dpVersion:((ce=P.version)==null?void 0:ce.kumaDp.version)||n.dpVersion,envoyVersion:((ue=P.version)==null?void 0:ue.envoy.version)||n.envoyVersion,selectedTime:n.selectedTime,selectedUpdateTime:n.selectedUpdateTime,version:P.version||n.version}},Ee),$={name:y,nameRoute:o,mesh:s,meshRoute:u,zone:W!=null?W:"\u2014",service:w!=null?w:"\u2014",serviceInsightRoute:ae,protocol:K!=null?K:"\u2014",status:Ae,totalUpdates:_.totalUpdates,totalRejectedUpdates:_.totalRejectedUpdates,dpVersion:(oe=_.dpVersion)!=null?oe:"\u2014",envoyVersion:(ie=_.envoyVersion)!=null?ie:"\u2014",warnings:[],unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:_.selectedUpdateTime?ye(new Date(_.selectedUpdateTime).toUTCString()):"\u2014",lastConnected:_.selectedTime?ye(new Date(_.selectedTime).toUTCString()):"\u2014",type:Ye(t.dataplane)};if(_.version){const{kind:n}=Ze(_.version);switch(n!==Fe&&$.warnings.push(n),n){case Qe:$.unsupportedEnvoyVersion=!0;break;case Je:$.unsupportedKumaDPVersion=!0;break}}return N.value&&_.dpVersion&&R.find(P=>P.label===We)&&typeof _.version.kumaDp.kumaCpCompatible=="boolean"&&!_.version.kumaDp.kumaCpCompatible&&($.warnings.push(Xe),$.kumaDpAndKumaCpMismatch=!0),$}async function J(t){var o;L.value=!0,k.value=t,ee("offset",t>0?t:null);const s=E.params.mesh,y=h;try{const{items:u,next:H}=await xe.getAllDataplaneOverviewsFromMesh({mesh:s},{size:y,offset:t});if(Array.isArray(u)&&u.length>0){u.sort(function(w,K){return w.name===K.name?w.mesh>K.mesh?1:-1:w.name.localeCompare(K.name)}),c.value=H,l.value=u,Q((o=m.name)!=null?o:u[0].name);const R=await Promise.all(l.value.map(w=>Se(w)));f.value.data=R,r.value=!1,I.value=!1}else Q(null),f.value.data=[],r.value=!0,I.value=!0}catch(u){u instanceof Error?p.value=u:console.error(u),I.value=!0}finally{L.value=!1}}function Q(t){var s;t&&l.value.length>0?(D.value=(s=l.value.find(y=>y.name===t))!=null?s:l.value[0],ee("name",D.value.name)):(D.value=null,ee("name",null))}return(t,s)=>(i(),B(et,null,{content:S(()=>{var y;return[O(tt,{"selected-entity-name":(y=D.value)==null?void 0:y.name,"page-size":h,"is-loading":L.value,error:p.value,"empty-state":Ve(),"table-data":d(Y),"table-data-is-empty":r.value,"show-details":"",next:c.value!==null,"page-offset":k.value,onTableAction:s[1]||(s[1]=o=>Q(o.name)),onLoadData:s[2]||(s[2]=o=>J(o))},{additionalControls:S(()=>[e("div",null,[Et,Me(e("select",{id:"data-planes-type-filter","onUpdate:modelValue":s[0]||(s[0]=o=>b.value=o),"data-testid":"data-planes-type-filter"},[(i(),v(A,null,M(F,(o,u)=>e("option",{key:u,value:o},g(o),9,It)),64))],512),[[ze,b.value]])]),O(d(je),{label:"Columns",icon:"cogwheel","button-appearance":"outline"},{items:S(()=>[e("div",{onClick:Ce},[(i(!0),v(A,null,M(d(Z),(o,u)=>(i(),B(d(He),{key:u,class:"table-header-selector-item",item:o},{default:S(()=>[e("label",{for:`data-plane-table-header-checkbox-${u}`,class:"k-checkbox table-header-selector-item-checkbox"},[e("input",{id:`data-plane-table-header-checkbox-${u}`,checked:o.isChecked,type:"checkbox",class:"k-input",onChange:H=>Pe(H,o.tableHeaderKey)},null,40,Ot),G(" "+g(o.label),1)],8,Kt)]),_:2},1032,["item"]))),128))])]),_:1}),O(d(fe),{class:"add-dp-button",appearance:"primary",to:d(q),"data-testid":"data-plane-create-data-plane-button",onClick:Ue},{default:S(()=>[Lt,G(" Create data plane proxy ")]),_:1},8,["to"]),d(E).query.ns?(i(),B(d(fe),{key:0,appearance:"primary",to:{name:"data-plane-list-view"},"data-testid":"data-plane-ns-back-button"},{default:S(()=>[Nt,G(" View All ")]),_:1})):x("",!0)]),_:1},8,["selected-entity-name","is-loading","error","empty-state","table-data","table-data-is-empty","next","page-offset"])]}),sidebar:S(()=>[D.value!==null?(i(),B(Ut,{key:0,"data-plane-overview":D.value},null,8,["data-plane-overview"])):(i(),B(st,{key:1}))]),_:1}))}});const Ft=De(Rt,[["__scopeId","data-v-ff8c3ecb"]]);export{Ft as default};
