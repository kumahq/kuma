import{d as te,cz as Le,e as m,f as e,E as ye,m as u,n as N,o,_ as ae,j as P,cA as De,a4 as fe,cB as we,cC as Ne,cD as Re,a as j,w as L,b as B,t as D,p as xe,F as K,h as R,c as z,C as Me,r as ze,x as Te,y as Ce,u as je,i as C,k as He,S as he,K as ge,a2 as be,cE as Be,cF as qe,cG as Ge,a7 as Ye,a8 as Fe,cH as Ze,a6 as Qe,g as We,cI as Je,cJ as Xe,cK as et,z as ke,q as tt,A as at,B as nt}from"./index.8c6a97c0.js";import{D as st}from"./DataOverview.7798887e.js";import{E as lt}from"./EntityTag.bfac5eb3.js";import{Y as ot}from"./YamlView.b86fd036.js";import{_ as it}from"./EmptyBlock.vue_vue_type_script_setup_true_lang.80be7515.js";import"./ErrorBlock.d4c6dfc2.js";import"./LoadingBlock.vue_vue_type_script_setup_true_lang.1c68a0d8.js";import"./index.58caa11d.js";import"./CodeBlock.31e0047b.js";import"./_commonjsHelpers.712cc82f.js";const rt={class:"content-wrapper"},dt={class:"content-wrapper__content component-frame"},ct={key:0,class:"content-wrapper__sidebar component-frame"},ut=te({__name:"ContentWrapper",setup(a){const p=Le();return(v,I)=>(o(),m("div",rt,[e("div",dt,[ye(v.$slots,"content",{},void 0,!0)]),u(p).sidebar?(o(),m("div",ct,[ye(v.$slots,"sidebar",{},void 0,!0)])):N("",!0)]))}});const pt=ae(ut,[["__scopeId","data-v-154249ab"]]),O=a=>(Te("data-v-5785bb6c"),a=a(),Ce(),a),mt={class:"entity-summary entity-section-list"},vt={class:"entity-title"},_t=O(()=>e("span",{class:"kutil-sr-only"},"Data plane:",-1)),yt={class:"definition"},ft=O(()=>e("span",null,"Mesh:",-1)),ht={key:0},gt=O(()=>e("h4",null,"Tags",-1)),bt={class:"tag-list"},kt={key:1},Dt=O(()=>e("h4",null,"Dependencies",-1)),wt={class:"mt-2 heading-with-icon"},Tt=O(()=>e("h4",null,"Insights",-1)),Ct={class:"entity-section-list"},Pt=["data-testid"],St=O(()=>e("span",null,"Connect time:",-1)),Ut=["data-testid"],Et=O(()=>e("span",null,"Disconnect time:",-1)),Vt={class:"definition"},It=O(()=>e("span",null,"Control plane instance ID:",-1)),At={key:0},Kt=O(()=>e("summary",null," Responses (acknowledged / sent) ",-1)),Ot=["data-testid"],$t=te({__name:"DataPlaneEntitySummary",props:{dataPlaneOverview:{type:Object,required:!0}},setup(a){const p=a,v={"Partially degraded":"partially_degraded",Offline:"offline",Online:"online"},I=P(()=>{const{name:r,mesh:c,dataplane:y}=p.dataPlaneOverview;return{type:"Dataplane",name:r,mesh:c,networking:y.networking}}),$=P(()=>De(p.dataPlaneOverview.dataplane)),S=P(()=>{const r=Array.from(p.dataPlaneOverview.dataplaneInsight.subscriptions);return r.reverse(),r.map(c=>{const y=c.connectTime!==void 0?fe(c.connectTime):"\u2014",i=c.disconnectTime!==void 0?fe(c.disconnectTime):"\u2014",d=Object.entries(c.status).filter(([g])=>!["total","lastUpdateTime"].includes(g)).map(([g,b])=>{var q,x,G,Y,F;const w=`${(q=b.responsesAcknowledged)!=null?q:0} / ${(x=b.responsesSent)!=null?x:0}`;return{type:g.toUpperCase(),ratio:w,responsesSent:(G=b.responsesSent)!=null?G:0,responsesAcknowledged:(Y=b.responsesAcknowledged)!=null?Y:0,responsesRejected:(F=b.responsesRejected)!=null?F:0}});return{subscription:c,formattedConnectDate:y,formattedDisconnectDate:i,statuses:d}})}),f=P(()=>{const{status:r}=we(p.dataPlaneOverview.dataplane,p.dataPlaneOverview.dataplaneInsight);return Ne[v[r]]}),U=P(()=>{const r=Re(p.dataPlaneOverview.dataplaneInsight);return r!==null?Object.entries(r).map(([c,y])=>({name:c,version:y})):[]}),h=P(()=>{const{subscriptions:r}=p.dataPlaneOverview.dataplaneInsight;if(r.length===0)return[];const c=[],y=r[r.length-1],i=y.version.envoy,d=y.version.kumaDp,g=i.kumaDpCompatible!==void 0?i.kumaDpCompatible:!0,b=d.kumaCpCompatible!==void 0?d.kumaCpCompatible:!0;if(!g){const w=`Envoy ${i.version} is not supported by Kuma DP ${d.version}.`;c.push(w)}if(!b){const w=`Kuma DP ${d.version} is not supported by this Kuma control plane.`;c.push(w)}return c});return(r,c)=>{const y=ze("router-link");return o(),m("div",mt,[e("section",null,[e("h3",vt,[_t,j(y,{to:{name:"data-plane-detail-view",params:{mesh:a.dataPlaneOverview.mesh,dataPlane:a.dataPlaneOverview.name}}},{default:L(()=>[B(D(a.dataPlaneOverview.name),1)]),_:1},8,["to"]),e("div",{class:xe(`status status--${u(f).appearance}`),"data-testid":"data-plane-status-badge"},D(u(f).title.toLowerCase()),3)]),e("div",yt,[ft,e("span",null,D(a.dataPlaneOverview.mesh),1)])]),u($).length>0?(o(),m("section",ht,[gt,e("div",bt,[(o(!0),m(K,null,R(u($),(i,d)=>(o(),z(lt,{key:d,tag:i},null,8,["tag"]))),128))])])):N("",!0),u(U).length>0?(o(),m("section",kt,[Dt,(o(!0),m(K,null,R(u(U),(i,d)=>(o(),m("div",{key:d,class:"definition"},[e("span",null,D(i.name)+":",1),e("span",null,D(i.version),1)]))),128)),u(h).length>0?(o(),m(K,{key:0},[e("h5",wt,[B(" Warnings "),j(u(Me),{class:"ml-1",icon:"warning",color:"var(--black-75)","secondary-color":"var(--yellow-300)",size:"20"})]),(o(!0),m(K,null,R(u(h),(i,d)=>(o(),m("p",{key:d},D(i),1))),128))],64)):N("",!0)])):N("",!0),u(S).length>0?(o(),m(K,{key:2},[e("section",null,[Tt,e("div",Ct,[(o(!0),m(K,null,R(u(S),(i,d)=>(o(),m("div",{key:d},[e("div",{class:"definition","data-testid":`data-plane-connect-time-${d}`},[St,e("span",null,D(i.formattedConnectDate),1)],8,Pt),e("div",{class:"definition","data-testid":`data-plane-disconnect-time-${d}`},[Et,e("span",null,D(i.formattedDisconnectDate),1)],8,Ut),e("div",Vt,[It,e("span",null,D(i.subscription.controlPlaneInstanceId),1)]),i.statuses.length>0?(o(),m("details",At,[Kt,(o(!0),m(K,null,R(i.statuses,(g,b)=>(o(),m("div",{key:`${d}-${b}`,class:"definition","data-testid":`data-plane-subscription-status-${d}-${b}`},[e("span",null,D(g.type)+":",1),e("span",null,D(g.ratio),1)],8,Ot))),128))])):N("",!0)]))),128))])]),e("section",null,[j(ot,{id:"code-block-data-plane-summary",content:u(I),"code-max-height":"250px"},null,8,["content"])])],64)):N("",!0)])}}});const Lt=ae($t,[["__scopeId","data-v-5785bb6c"]]),Pe=[{key:"status",label:"Status"},{key:"name",label:"Name"},{key:"mesh",label:"Mesh"},{key:"type",label:"Type"},{key:"service",label:"Service"},{key:"protocol",label:"Protocol"},{key:"zone",label:"Zone"},{key:"lastConnected",label:"Last Connected"},{key:"lastUpdated",label:"Last Updated"},{key:"totalUpdates",label:"Total Updates"},{key:"dpVersion",label:"Kuma DP version"},{key:"envoyVersion",label:"Envoy version"},{key:"details",label:"Details",hideLabel:!0}],Nt=["name","details"],Rt=Pe.filter(a=>!Nt.includes(a.key)).map(a=>({tableHeaderKey:a.key,label:a.label,isChecked:!1})),Se=["status","name","mesh","type","service","protocol","zone","lastUpdated","dpVersion","details"];function xt(a,p=Se){return Pe.filter(v=>p.includes(v.key)?a?!0:v.key!=="zone":!1)}function ee(a,p){const v=window.history.state;if(v===null)return;const I=v.current.indexOf("?"),$=I>-1?v.current.substring(I):"",S=new URLSearchParams($);p!=null?S.set(a,String(p)):S.has(a)&&S.delete(a);const f=S.toString(),U=f===""?"":"?"+f;let h="";if(I>-1?h=v.current.substring(0,I)+U:h=v.current+U,v.current!==h){const r=Object.assign({},v);r.current=h,window.history.replaceState(r,"","#"+h)}}const ne=a=>(Te("data-v-bee730dc"),a=a(),Ce(),a),Mt=ne(()=>e("label",{for:"data-planes-type-filter",class:"mr-2"}," Type: ",-1)),zt=["value"],jt=["for"],Ht=["id","checked","onChange"],Bt=ne(()=>e("span",{class:"custom-control-icon"}," + ",-1)),qt=ne(()=>e("span",{class:"custom-control-icon"}," \u2190 ",-1)),Gt=te({__name:"DataPlaneListView",props:{name:{type:String,required:!1,default:null},offset:{type:Number,required:!1,default:0}},setup(a){const p=a,v=50,I=["All","Standard","Gateway (builtin)","Gateway (provided)"],$=je(),S=tt(),f=C(Se),U=C(!0),h=C(!1),r=C(!1),c=C(!1),y=C({headers:[],data:[]}),i=C([]),d=C(null),g=C("All"),b=C(p.offset),w=C(null),q=P(()=>S.getters["config/getEnvironment"]),x=P(()=>S.getters["config/getMulticlusterStatus"]),G=P(()=>q.value==="universal"?{name:"universal-dataplane"}:{name:"kubernetes-dataplane"}),Y=P(()=>{const t=y.value.data.filter(_=>g.value==="All"?!0:_.type.toLowerCase()===g.value.toLowerCase()),l=xt(x.value,f.value);return{data:t,headers:l}}),F=P(()=>Rt.filter(t=>x.value?!0:t.tableHeaderKey!=="zone").map(t=>{const l=f.value.includes(t.tableHeaderKey);return{...t,isChecked:l}}));He(()=>$.params.mesh,function(){$.name==="data-plane-list-view"&&(U.value=!0,h.value=!1,r.value=!1,c.value=!1,Z(0))});const se=he.get("dpVisibleTableHeaderKeys");Array.isArray(se)&&(f.value=se),Z(p.offset);function Ue(t){t.stopPropagation()}function Ee(t,l){const _=t.target,n=f.value.findIndex(T=>T===l);_.checked&&n===-1?f.value.push(l):!_.checked&&n>-1&&f.value.splice(n,1),he.set("dpVisibleTableHeaderKeys",Array.from(new Set(f.value)))}function Ve(){at.logger.info(nt.CREATE_DATA_PLANE_PROXY_CLICKED)}function Ie(){return{title:"No Data",message:"There are no data plane proxies present."}}async function Ae(t){var oe,ie,re,de,ce;const l=t.mesh,_=t.name,n={name:"data-plane-detail-view",params:{mesh:l,dataPlane:_}},T={name:"mesh-child",params:{mesh:l}},H=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],E=De(t.dataplane).filter(s=>H.includes(s.label)),A=(oe=E.find(s=>s.label==="kuma.io/service"))==null?void 0:oe.value,W=(ie=E.find(s=>s.label==="kuma.io/protocol"))==null?void 0:ie.value,J=(re=E.find(s=>s.label==="kuma.io/zone"))==null?void 0:re.value;let le;A!==void 0&&(le={name:"service-insight-detail-view",params:{mesh:l,service:A}});const{status:Oe}=we(t.dataplane,t.dataplaneInsight),$e={totalUpdates:0,totalRejectedUpdates:0,dpVersion:null,envoyVersion:null,selectedTime:NaN,selectedUpdateTime:NaN,version:null},k=t.dataplaneInsight.subscriptions.reduce((s,V)=>{var ue,pe,me,ve;if(V.connectTime){const _e=Date.parse(V.connectTime);(!s.selectedTime||_e>s.selectedTime)&&(s.selectedTime=_e)}const X=Date.parse(V.status.lastUpdateTime);return X&&(!s.selectedUpdateTime||X>s.selectedUpdateTime)&&(s.selectedUpdateTime=X),{totalUpdates:s.totalUpdates+parseInt((ue=V.status.total.responsesSent)!=null?ue:"0",10),totalRejectedUpdates:s.totalRejectedUpdates+parseInt((pe=V.status.total.responsesRejected)!=null?pe:"0",10),dpVersion:((me=V.version)==null?void 0:me.kumaDp.version)||s.dpVersion,envoyVersion:((ve=V.version)==null?void 0:ve.envoy.version)||s.envoyVersion,selectedTime:s.selectedTime,selectedUpdateTime:s.selectedUpdateTime,version:V.version||s.version}},$e),M={name:_,nameRoute:n,mesh:l,meshRoute:T,zone:J!=null?J:"\u2014",service:A!=null?A:"\u2014",serviceInsightRoute:le,protocol:W!=null?W:"\u2014",status:Oe,totalUpdates:k.totalUpdates,totalRejectedUpdates:k.totalRejectedUpdates,dpVersion:(de=k.dpVersion)!=null?de:"\u2014",envoyVersion:(ce=k.envoyVersion)!=null?ce:"\u2014",warnings:[],unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:k.selectedUpdateTime?be(new Date(k.selectedUpdateTime).toUTCString()):"\u2014",lastConnected:k.selectedTime?be(new Date(k.selectedTime).toUTCString()):"\u2014",type:Be(t.dataplane)};if(k.version){const{kind:s}=qe(k.version);switch(s!==Ge&&M.warnings.push(s),s){case Fe:M.unsupportedEnvoyVersion=!0;break;case Ye:M.unsupportedKumaDPVersion=!0;break}}return x.value&&k.dpVersion&&E.find(V=>V.label===Ze)&&typeof k.version.kumaDp.kumaCpCompatible=="boolean"&&!k.version.kumaDp.kumaCpCompatible&&(M.warnings.push(Qe),M.kumaDpAndKumaCpMismatch=!0),M}async function Z(t){var _;U.value=!0,b.value=t,ee("offset",t>0?t:null);const l=$.params.mesh||null;try{const{items:n,next:T}=await Ke(l,v,t);if(n.length>0){n.sort(function(E,A){return E.name===A.name?E.mesh>A.mesh?1:-1:E.name.localeCompare(A.name)}),d.value=T,i.value=n,Q((_=p.name)!=null?_:n[0].name);const H=await Promise.all(i.value.map(E=>Ae(E)));y.value.data=H,c.value=!1,h.value=!1}else Q(null),y.value.data=[],c.value=!0,h.value=!0}catch(n){r.value=!0,h.value=!0,console.error(n)}finally{U.value=!1}}function Ke(t,l,_){return t==="all"||t===null?ge.getAllDataplaneOverviews({size:l,offset:_}):ge.getAllDataplaneOverviewsFromMesh({mesh:t})}function Q(t){var l;t&&i.value.length>0?(w.value=(l=i.value.find(_=>_.name===t))!=null?l:i.value[0],ee("name",w.value.name)):(w.value=null,ee("name",null))}return(t,l)=>(o(),z(pt,null,{content:L(()=>{var _;return[j(st,{"selected-entity-name":(_=w.value)==null?void 0:_.name,"page-size":v,"has-error":r.value,"is-loading":U.value,"empty-state":Ie(),"table-data":u(Y),"table-data-is-empty":c.value,"show-details":"",next:d.value!==null,"page-offset":b.value,onTableAction:l[1]||(l[1]=n=>Q(n.name)),onLoadData:l[2]||(l[2]=n=>Z(n))},{additionalControls:L(()=>[e("div",null,[Mt,We(e("select",{id:"data-planes-type-filter","onUpdate:modelValue":l[0]||(l[0]=n=>g.value=n),"data-testid":"data-planes-type-filter"},[(o(),m(K,null,R(I,(n,T)=>e("option",{key:T,value:n},D(n),9,zt)),64))],512),[[Je,g.value]])]),j(u(Xe),{label:"Columns",icon:"cogwheel","button-appearance":"outline"},{items:L(()=>[e("div",{onClick:Ue},[(o(!0),m(K,null,R(u(F),(n,T)=>(o(),z(u(et),{key:T,class:"table-header-selector-item",item:n},{default:L(()=>[e("label",{for:`data-plane-table-header-checkbox-${T}`,class:"k-checkbox table-header-selector-item-checkbox"},[e("input",{id:`data-plane-table-header-checkbox-${T}`,checked:n.isChecked,type:"checkbox",class:"k-input",onChange:H=>Ee(H,n.tableHeaderKey)},null,40,Ht),B(" "+D(n.label),1)],8,jt)]),_:2},1032,["item"]))),128))])]),_:1}),j(u(ke),{class:"add-dp-button",appearance:"primary",to:u(G),"data-testid":"data-plane-create-data-plane-button",onClick:Ve},{default:L(()=>[Bt,B(" Create data plane proxy ")]),_:1},8,["to"]),t.$route.query.ns?(o(),z(u(ke),{key:0,appearance:"primary",to:{name:"data-plane-list-view"},"data-testid":"data-plane-ns-back-button"},{default:L(()=>[qt,B(" View All ")]),_:1})):N("",!0)]),_:1},8,["selected-entity-name","has-error","is-loading","empty-state","table-data","table-data-is-empty","next","page-offset"])]}),sidebar:L(()=>[w.value!==null?(o(),z(Lt,{key:0,"data-plane-overview":w.value},null,8,["data-plane-overview"])):(o(),z(it,{key:1}))]),_:1}))}});const na=ae(Gt,[["__scopeId","data-v-bee730dc"]]);export{na as default};
