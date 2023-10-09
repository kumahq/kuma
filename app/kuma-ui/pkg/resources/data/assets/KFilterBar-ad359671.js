var ce=Object.defineProperty;var de=(n,o,a)=>o in n?ce(n,o,{enumerable:!0,configurable:!0,writable:!0,value:a}):n[o]=a;var Q=(n,o,a)=>(de(n,typeof o!="symbol"?o+"":o,a),a);import{d as oe,a2 as pe,L as me,r as fe,o as p,g as B,w as y,R as ie,h as D,l as f,D as b,j as k,F as L,U as ge,i as c,a9 as ve,m as w,G as le,k as O,s as ye,K as M,$ as he,W as be,a0 as ke,a1 as _e,Z as Te,q as re,v as z,f as V,ar as ae,as as Se,at as we,au as Ce,y as ne,av as xe,aw as De,x as Ie,S as Ue,B as ze,C as Le}from"./index-35b33747.js";import{d as Ae,a as Ne,c as Fe,C as je}from"./dataplane-0a086c06.js";const Ee={key:0},Be=oe({__name:"DataPlaneList",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{},gateways:{type:Boolean,default:!1}},emits:["load-data","change"],setup(n,{emit:o}){const a=n,{t:i,formatIsoDate:s}=pe(),g=me()("use zones");function T(v){return v.map(l=>{var j,I,K,e,r,m,U,Y;const S=l.mesh,t=l.name,_=((j=l.dataplane.networking.gateway)==null?void 0:j.type)||"STANDARD",x={name:_==="STANDARD"?"data-plane-detail-view":"gateway-detail-view",params:{mesh:S,dataPlane:t}},H=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],A=Ae(l.dataplane).filter(u=>H.includes(u.label)),E=(I=A.find(u=>u.label==="kuma.io/service"))==null?void 0:I.value,Z=(K=A.find(u=>u.label==="kuma.io/protocol"))==null?void 0:K.value,N=(e=A.find(u=>u.label==="kuma.io/zone"))==null?void 0:e.value;let $;E!==void 0&&($={name:"service-detail-view",params:{mesh:S,service:E}});let R;N!==void 0&&(R={name:"zone-cp-detail-view",params:{zone:N}});const{status:q}=Ne(l.dataplane,l.dataplaneInsight),W=((r=l.dataplaneInsight)==null?void 0:r.subscriptions)??[],G={totalUpdates:0,totalRejectedUpdates:0,dpVersion:null,envoyVersion:null,selectedTime:NaN,selectedUpdateTime:NaN,version:null},h=W.reduce((u,C)=>{var X,ee;if(C.connectTime){const te=Date.parse(C.connectTime);(!u.selectedTime||te>u.selectedTime)&&(u.selectedTime=te)}const J=Date.parse(C.status.lastUpdateTime);return J&&(!u.selectedUpdateTime||J>u.selectedUpdateTime)&&(u.selectedUpdateTime=J),{totalUpdates:u.totalUpdates+parseInt(C.status.total.responsesSent??"0",10),totalRejectedUpdates:u.totalRejectedUpdates+parseInt(C.status.total.responsesRejected??"0",10),dpVersion:((X=C.version)==null?void 0:X.kumaDp.version)||u.dpVersion,envoyVersion:((ee=C.version)==null?void 0:ee.envoy.version)||u.envoyVersion,selectedTime:u.selectedTime,selectedUpdateTime:u.selectedUpdateTime,version:C.version||u.version}},G),F={name:t,dataplaneInsight:l.dataplaneInsight,detailViewRoute:x,type:_,zone:{title:N??i("common.collection.none"),route:R},service:{title:E??i("common.collection.none"),route:$},protocol:Z??i("common.collection.none"),status:q,totalUpdates:h.totalUpdates,totalRejectedUpdates:h.totalRejectedUpdates,envoyVersion:h.envoyVersion??i("common.collection.none"),warnings:{version_mismatch:!1,cert_expired:!1},lastUpdated:h.selectedUpdateTime?s(new Date(h.selectedUpdateTime).toUTCString()):i("common.collection.none"),lastConnected:h.selectedTime?s(new Date(h.selectedTime).toUTCString()):i("common.collection.none"),overview:l};if(h.version){const{kind:u}=Fe(h.version);u!==je&&(F.warnings.version_mismatch=!0)}g&&h.dpVersion&&A.find(C=>C.label==="kuma.io/zone")&&typeof((m=h.version)==null?void 0:m.kumaDp.kumaCpCompatible)=="boolean"&&!h.version.kumaDp.kumaCpCompatible&&(F.warnings.version_mismatch=!0);const P=(Y=(U=l.dataplaneInsight)==null?void 0:U.mTLS)==null?void 0:Y.certificateExpirationTime;return P&&Date.now()>new Date(P).getTime()&&(F.warnings.cert_expired=!0),F})}return(v,l)=>{const S=fe("RouterLink");return p(),B(Te,{"empty-state-message":c(i)("common.emptyState.message",{type:a.gateways?"Gateways":"Data Plane Proxies"}),"empty-state-cta-to":c(i)(`data-planes.href.docs.${a.gateways?"gateway":"data_plane_proxy"}`),"empty-state-cta-text":c(i)("common.documentation"),headers:[{label:"Name",key:"name"},...a.gateways?[{label:"Type",key:"type"}]:[],{label:"Service",key:"service"},...a.gateways?[]:[{label:"Protocol",key:"protocol"}],...c(g)?[{label:"Zone",key:"zone"}]:[],{label:"Last Updated",key:"lastUpdated"},{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}],"page-number":a.pageNumber,"page-size":a.pageSize,total:a.total,items:a.items?T(a.items):void 0,error:a.error,onChange:l[0]||(l[0]=t=>o("change",t))},{toolbar:y(()=>[ie(v.$slots,"toolbar",{},void 0,!0)]),name:y(({row:t})=>[D(S,{to:{name:a.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:t.name}},"data-testid":"detail-view-link"},{default:y(()=>[f(b(t.name),1)]),_:2},1032,["to"])]),service:y(({rowValue:t})=>[t.route?(p(),B(S,{key:0,to:t.route},{default:y(()=>[f(b(t.title),1)]),_:2},1032,["to"])):(p(),k(L,{key:1},[f(b(t.title),1)],64))]),zone:y(({rowValue:t})=>[t.route?(p(),B(S,{key:0,to:t.route},{default:y(()=>[f(b(t.title),1)]),_:2},1032,["to"])):(p(),k(L,{key:1},[f(b(t.title),1)],64))]),status:y(({rowValue:t})=>[t?(p(),B(ge,{key:0,status:t},null,8,["status"])):(p(),k(L,{key:1},[f(b(c(i)("common.collection.none")),1)],64))]),warnings:y(({row:t})=>[Object.values(t.warnings).some(_=>_)?(p(),B(c(ve),{key:0},{content:y(()=>[w("ul",null,[(p(!0),k(L,null,le(t.warnings,(_,x)=>(p(),k(L,{key:x},[_?(p(),k("li",Ee,b(c(i)(`data-planes.components.data-plane-list.${x}`)),1)):O("",!0)],64))),128))])]),default:y(()=>[f(),D(ye,{class:"mr-1",size:c(M),"hide-title":""},null,8,["size"])]),_:2},1024)):(p(),k(L,{key:1},[f(b(c(i)("common.collection.none")),1)],64))]),certificate:y(({row:t})=>{var _,x;return[f(b((x=(_=t.dataplaneInsight)==null?void 0:_.mTLS)!=null&&x.certificateExpirationTime?c(s)(new Date(t.dataplaneInsight.mTLS.certificateExpirationTime).toUTCString()):c(i)("data-planes.components.data-plane-list.certificate.none")),1)]}),actions:y(({row:t})=>[D(c(he),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:y(()=>[D(c(be),{class:"non-visual-button",appearance:"secondary",size:"small"},{default:y(()=>[D(c(ke),{size:c(M)},null,8,["size"])]),_:1})]),items:y(()=>[D(c(_e),{item:{to:{name:a.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:t.name}},label:c(i)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:3},8,["empty-state-message","empty-state-cta-to","empty-state-cta-text","headers","page-number","page-size","total","items","error"])}}});const lt=re(Be,[["__scopeId","data-v-107acb3a"]]);function Me(n,o,a){return Math.max(o,Math.min(n,a))}const $e=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class Re{constructor(o,a){Q(this,"commands");Q(this,"keyMap");Q(this,"boundTriggerShortcuts");this.commands=a,this.keyMap=Object.fromEntries(Object.entries(o).map(([i,s])=>[i.toLowerCase(),s])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(o){qe(o,this.keyMap,this.commands)}}function qe(n,o,a){const i=Pe(n.code),s=[n.ctrlKey?"ctrl":"",n.shiftKey?"shift":"",n.altKey?"alt":"",i].filter(T=>T!=="").join("+"),d=o[s];if(!d)return;const g=a[d];g.isAllowedContext&&!g.isAllowedContext(n)||(g.shouldPreventDefaultAction&&n.preventDefault(),!(g.isDisabled&&g.isDisabled())&&g.trigger(n))}function Pe(n){return $e.includes(n)?"":n.replace(/^Key/,"").toLowerCase()}function Ke(n,o){const a=" "+n,i=a.matchAll(/ ([-\s\w]+):\s*/g),s=[];for(const d of Array.from(i)){if(d.index===void 0)continue;const g=Qe(d[1]);if(o.length>0&&!o.includes(g))throw new Error(`Unknown field “${g}”. Known fields: ${o.join(", ")}`);const T=d.index+d[0].length,v=a.substring(T);let l;if(/^\s*["']/.test(v)){const t=v.match(/['"](.*?)['"]/);if(t!==null)l=t[1];else throw new Error(`Quote mismatch for field “${g}”.`)}else{const t=v.indexOf(" "),_=t===-1?v.length:t;l=v.substring(0,_)}l!==""&&s.push([g,l])}return s}function Qe(n){return n.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(o,a)=>a===0?o:o.substring(1).toUpperCase())}let se=0;const Ve=(n="unique")=>(se++,`${n}-${se}`),ue=n=>(ze("data-v-9e2bf5f8"),n=n(),Le(),n),Oe=ue(()=>w("span",{class:"visually-hidden"},"Focus filter",-1)),He={class:"k-filter-icon"},Ze=["for"],We=["id","placeholder"],Ge={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},Je={class:"k-suggestion-list"},Ye={key:0,class:"k-filter-bar-error"},Xe={key:0},et=["title","data-filter-field"],tt={class:"visually-hidden"},at=ue(()=>w("span",{class:"visually-hidden"},"Clear query",-1)),nt=oe({__name:"KFilterBar",props:{id:{type:String,required:!1,default:()=>Ve("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(n,{emit:o}){const a=n,i=z(null),s=z(null),d=z(a.query),g=z([]),T=z(null),v=z(!1),l=z(-1),S=V(()=>Object.keys(a.fields)),t=V(()=>Object.entries(a.fields).slice(0,5).map(([e,r])=>({fieldName:e,...r}))),_=V(()=>S.value.length>0?`Filter by ${S.value.join(", ")}`:"Filter"),x=V(()=>a.placeholder??_.value);ae(()=>g.value,function(e,r){K(e,r)||(T.value=null,o("fields-change",{fields:e,query:d.value}))}),ae(()=>d.value,function(){d.value===""&&(T.value=null),v.value=!0});const H={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},A={submitQuery:{trigger:N,isAllowedContext(e){return s.value!==null&&e.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:$,isAllowedContext(e){return s.value!==null&&e.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:R,isAllowedContext(e){return s.value!==null&&e.composedPath().includes(s.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:j,isAllowedContext(e){return i.value!==null&&e.composedPath().includes(i.value)}}};function E(){const e=new Re(H,A);Ie(function(){e.registerListener()}),Ue(function(){e.unRegisterListener()}),I(d.value)}E();function Z(e){const r=e.target;I(r.value)}function N(){if(s.value instanceof HTMLInputElement)if(l.value===-1)I(s.value.value),v.value=!1;else{const e=t.value[l.value].fieldName;e&&h(s.value,e)}}function $(){q(1)}function R(){q(-1)}function q(e){l.value=Me(l.value+e,-1,t.value.length-1)}function W(){s.value instanceof HTMLInputElement&&s.value.focus()}function G(e){const m=e.currentTarget.getAttribute("data-filter-field");m&&s.value instanceof HTMLInputElement&&h(s.value,m)}function h(e,r){const m=d.value===""||d.value.endsWith(" ")?"":" ";d.value+=m+r+":",e.focus(),l.value=-1}function F(){d.value="",s.value instanceof HTMLInputElement&&(s.value.value="",s.value.focus(),I(""))}function P(e){e.relatedTarget===null&&j(),i.value instanceof HTMLElement&&e.relatedTarget instanceof Node&&!i.value.contains(e.relatedTarget)&&j()}function j(){v.value=!1}function I(e){T.value=null;try{const r=Ke(e,S.value);r.sort((m,U)=>m[0].localeCompare(U[0])),g.value=r}catch(r){if(r instanceof Error)T.value=r,v.value=!0;else throw r}}function K(e,r){return JSON.stringify(e)===JSON.stringify(r)}return(e,r)=>(p(),k("div",{ref_key:"filterBar",ref:i,class:"k-filter-bar","data-testid":"k-filter-bar"},[w("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:W},[Oe,f(),w("span",He,[D(c(Se),{decorative:"","data-testid":"k-filter-bar-filter-icon","hide-title":"",size:c(M)},null,8,["size"])])]),f(),w("label",{for:`${a.id}-filter-bar-input`,class:"visually-hidden"},[ie(e.$slots,"default",{},()=>[f(b(_.value),1)],!0)],8,Ze),f(),we(w("input",{id:`${a.id}-filter-bar-input`,ref_key:"filterInput",ref:s,"onUpdate:modelValue":r[0]||(r[0]=m=>d.value=m),class:"k-filter-bar-input",type:"text",placeholder:x.value,"data-testid":"k-filter-bar-filter-input",onFocus:r[1]||(r[1]=m=>v.value=!0),onBlur:P,onChange:Z},null,40,We),[[Ce,d.value]]),f(),v.value?(p(),k("div",Ge,[w("div",Je,[T.value!==null?(p(),k("p",Ye,b(T.value.message),1)):(p(),k("button",{key:1,class:ne(["k-submit-query-button",{"k-submit-query-button-is-selected":l.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:N},`
          Submit `+b(d.value),3)),f(),(p(!0),k(L,null,le(t.value,(m,U)=>(p(),k("div",{key:`${a.id}-${U}`,class:ne(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":l.value===U}])},[w("b",null,b(m.fieldName),1),m.description!==""?(p(),k("span",Xe,": "+b(m.description),1)):O("",!0),f(),w("button",{class:"k-apply-suggestion-button",title:`Add ${m.fieldName}:`,type:"button","data-filter-field":m.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:G},[w("span",tt,"Add "+b(m.fieldName)+":",1),f(),D(c(xe),{decorative:"","hide-title":"",size:c(M)},null,8,["size"])],8,et)],2))),128))])])):O("",!0),f(),d.value!==""?(p(),k("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:F},[at,f(),D(c(De),{decorative:"","hide-title":"",size:c(M)},null,8,["size"])])):O("",!0)],512))}});const rt=re(nt,[["__scopeId","data-v-9e2bf5f8"]]);export{lt as D,rt as K};
