var ie=Object.defineProperty;var oe=(s,a,t)=>a in s?ie(s,a,{enumerable:!0,configurable:!0,writable:!0,value:t}):s[a]=t;var U=(s,a,t)=>(oe(s,typeof a!="symbol"?a+"":a,t),t);import{K as q}from"./index-9dd3e7d3.js";import{d as te,l as le,a as O,o as c,b as H,a0 as re,w as h,r as se,e as I,f,t as _,c as g,F as N,C as Y,q as m,p as S,v as Z,V as ue,_ as ne,m as F,O as J,aa as G,as as ce,at as de,au as pe,n as X,av as fe,aw as me,I as ge,T as ve,A as he,B as ye}from"./index-203d56a2.js";import{A as be}from"./AppCollection-0aeb195e.js";import{S as _e}from"./StatusBadge-01928c30.js";import{_ as ke}from"./WarningIcon.vue_vue_type_script_setup_true_lang-297a38e2.js";import{d as Se,a as Ce,c as Te,C as we}from"./dataplane-dcd0858b.js";const xe={key:0},Ie=te({__name:"DataPlaneList",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{},isSelectedRow:{type:[Function,null],default:null},summaryRouteName:{},canUseZones:{type:Boolean}},emits:["change"],setup(s,{emit:a}){const{t,formatIsoDate:w}=le(),l=s,n=a;function i(y){return y.map(o=>{var M,B,K,j,E,P,V;const{mesh:d,name:b}=o,v=(M=o.dataplane.networking.gateway)!=null&&M.type?t(`data-planes.type.${o.dataplane.networking.gateway.type.toLowerCase()}`):t("data-planes.type.standard"),C=Se(o.dataplane),u=C.filter(k=>k.label==="kuma.io/service"),{status:T}=Ce(o.dataplane,o.dataplaneInsight),L=((B=o.dataplaneInsight)==null?void 0:B.subscriptions)??[],W={dpVersion:null,version:null},A=L.reduce((k,x)=>{var R;return{dpVersion:((R=x.version)==null?void 0:R.kumaDp.version)||k.dpVersion,version:x.version||k.version}},W);let z;(j=(K=o.dataplaneInsight)==null?void 0:K.mTLS)!=null&&j.certificateExpirationTime?z=w(o.dataplaneInsight.mTLS.certificateExpirationTime):z=t("data-planes.components.data-plane-list.certificate.none");const D={name:b,type:v,mesh:d,services:u,status:T,warnings:{version_mismatch:!1,cert_expired:!1},certificate:z};if(A.version){const{kind:k}=Te(A.version);k!==we&&(D.warnings.version_mismatch=!0)}l.canUseZones&&A.dpVersion&&C.find(x=>x.label==="kuma.io/zone")&&typeof((E=A.version)==null?void 0:E.kumaDp.kumaCpCompatible)=="boolean"&&!A.version.kumaDp.kumaCpCompatible&&(D.warnings.version_mismatch=!0);const $=(V=(P=o.dataplaneInsight)==null?void 0:P.mTLS)==null?void 0:V.certificateExpirationTime;return $&&Date.now()>new Date($).getTime()&&(D.warnings.cert_expired=!0),D})}return(y,o)=>{const d=O("RouterLink"),b=O("KBadge"),v=O("KTruncate"),C=O("KTooltip");return c(),H(be,{"empty-state-message":m(t)("common.emptyState.message",{type:"Data Plane Proxies"}),"empty-state-cta-to":m(t)("data-planes.href.docs.data_plane_proxy"),"empty-state-cta-text":m(t)("common.documentation"),headers:[{label:"Name",key:"name"},{label:"Type",key:"type"},{label:"Services",key:"services"},{label:"Certificate Info",key:"certificate"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Details",key:"details",hideLabel:!0}],"page-number":l.pageNumber,"page-size":l.pageSize,total:l.total,items:l.items?i(l.items):void 0,error:l.error,"is-selected-row":l.isSelectedRow,onChange:o[0]||(o[0]=u=>n("change",u))},re({name:h(({row:u})=>[I(d,{to:{name:l.summaryRouteName,params:{mesh:u.mesh,dataPlane:u.name},query:{page:l.pageNumber,size:l.pageSize}}},{default:h(()=>[f(_(u.name),1)]),_:2},1032,["to"])]),services:h(({row:u})=>[u.services.length>0?(c(),H(v,{key:0},{default:h(()=>[(c(!0),g(N,null,Y(u.services,(T,L)=>(c(),H(b,{key:L},{default:h(()=>[I(d,{to:{name:"service-detail-view",params:{mesh:u.mesh,service:T.value}}},{default:h(()=>[f(_(T.value),1)]),_:2},1032,["to"])]),_:2},1024))),128))]),_:2},1024)):(c(),g(N,{key:1},[f(_(m(t)("common.collection.none")),1)],64))]),status:h(({row:u})=>[I(_e,{status:u.status},null,8,["status"])]),warnings:h(({row:u})=>[Object.values(u.warnings).some(T=>T)?(c(),H(C,{key:0},{content:h(()=>[S("ul",null,[(c(!0),g(N,null,Y(u.warnings,(T,L)=>(c(),g(N,{key:L},[T?(c(),g("li",xe,_(m(t)(`data-planes.components.data-plane-list.${L}`)),1)):Z("",!0)],64))),128))])]),default:h(()=>[f(),I(ke,{class:"mr-1",size:m(q),"hide-title":""},null,8,["size"])]),_:2},1024)):(c(),g(N,{key:1},[f(_(m(t)("common.collection.none")),1)],64))]),details:h(({row:u})=>[I(d,{class:"details-link","data-testid":"details-link",to:{name:"data-plane-detail-view",params:{dataPlane:u.name}}},{default:h(()=>[f(_(m(t)("common.collection.details_link"))+" ",1),I(m(ue),{display:"inline-block",decorative:"",size:m(q)},null,8,["size"])]),_:2},1032,["to"])]),_:2},[y.$slots.toolbar?{name:"toolbar",fn:h(()=>[se(y.$slots,"toolbar",{},void 0,!0)]),key:"0"}:void 0]),1032,["empty-state-message","empty-state-cta-to","empty-state-cta-text","page-number","page-size","total","items","error","is-selected-row"])}}});const tt=ne(Ie,[["__scopeId","data-v-785a5e8d"]]);function Le(s,a,t){return Math.max(a,Math.min(s,t))}const Ae=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class De{constructor(a,t){U(this,"commands");U(this,"keyMap");U(this,"boundTriggerShortcuts");this.commands=t,this.keyMap=Object.fromEntries(Object.entries(a).map(([w,l])=>[w.toLowerCase(),l])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(a){Fe(a,this.keyMap,this.commands)}}function Fe(s,a,t){const w=Ne(s.code),l=[s.ctrlKey?"ctrl":"",s.shiftKey?"shift":"",s.altKey?"alt":"",w].filter(y=>y!=="").join("+"),n=a[l];if(!n)return;const i=t[n];i.isAllowedContext&&!i.isAllowedContext(s)||(i.shouldPreventDefaultAction&&s.preventDefault(),!(i.isDisabled&&i.isDisabled())&&i.trigger(s))}function Ne(s){return Ae.includes(s)?"":s.replace(/^Key/,"").toLowerCase()}function ze(s,a){const t=" "+s,w=t.matchAll(/ ([-\s\w]+):\s*/g),l=[];for(const n of Array.from(w)){if(n.index===void 0)continue;const i=Be(n[1]);if(a.length>0&&!a.includes(i))throw new Error(`Unknown field “${i}”. Known fields: ${a.join(", ")}`);const y=n.index+n[0].length,o=t.substring(y);let d;if(/^\s*["']/.test(o)){const v=o.match(/['"](.*?)['"]/);if(v!==null)d=v[1];else throw new Error(`Quote mismatch for field “${i}”.`)}else{const v=o.indexOf(" "),C=v===-1?o.length:v;d=o.substring(0,C)}d!==""&&l.push([i,d])}return l}function Be(s){return s.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(a,t)=>t===0?a:a.substring(1).toUpperCase())}let ee=0;const Ee=(s="unique")=>(ee++,`${s}-${ee}`),ae=s=>(he("data-v-349996e3"),s=s(),ye(),s),qe=ae(()=>S("span",{class:"visually-hidden"},"Focus filter",-1)),$e={class:"k-filter-icon"},Me=["for"],Ke=["id","placeholder"],je={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},Pe={class:"k-suggestion-list"},Ve={key:0,class:"k-filter-bar-error"},Re={key:0},Qe=["title","data-filter-field"],Ue={class:"visually-hidden"},Oe=ae(()=>S("span",{class:"visually-hidden"},"Clear query",-1)),He=te({__name:"FilterBar",props:{id:{type:String,required:!1,default:()=>Ee("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(s,{emit:a}){const t=s,w=a,l=F(null),n=F(null),i=F(t.query),y=F([]),o=F(null),d=F(!1),b=F(-1),v=J(()=>Object.keys(t.fields)),C=J(()=>Object.entries(t.fields).slice(0,5).map(([e,r])=>({fieldName:e,...r}))),u=J(()=>v.value.length>0?`Filter by ${v.value.join(", ")}`:"Filter"),T=J(()=>t.placeholder??u.value);G(()=>y.value,function(e,r){R(e,r)||(o.value=null,w("fields-change",{fields:e,query:i.value}))}),G(()=>i.value,function(){i.value===""&&(o.value=null),d.value=!0});const L={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},W={submitQuery:{trigger:D,isAllowedContext(e){return n.value!==null&&e.composedPath().includes(n.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:$,isAllowedContext(e){return n.value!==null&&e.composedPath().includes(n.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:M,isAllowedContext(e){return n.value!==null&&e.composedPath().includes(n.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:k,isAllowedContext(e){return l.value!==null&&e.composedPath().includes(l.value)}}};function A(){const e=new De(L,W);ge(function(){e.registerListener()}),ve(function(){e.unRegisterListener()}),x(i.value)}A();function z(e){const r=e.target;x(r.value)}function D(){if(n.value instanceof HTMLInputElement)if(b.value===-1)x(n.value.value),d.value=!1;else{const e=C.value[b.value].fieldName;e&&E(n.value,e)}}function $(){B(1)}function M(){B(-1)}function B(e){b.value=Le(b.value+e,-1,C.value.length-1)}function K(){n.value instanceof HTMLInputElement&&n.value.focus()}function j(e){const p=e.currentTarget.getAttribute("data-filter-field");p&&n.value instanceof HTMLInputElement&&E(n.value,p)}function E(e,r){const p=i.value===""||i.value.endsWith(" ")?"":" ";i.value+=p+r+":",e.focus(),b.value=-1}function P(){i.value="",n.value instanceof HTMLInputElement&&(n.value.value="",n.value.focus(),x(""))}function V(e){e.relatedTarget===null&&k(),l.value instanceof HTMLElement&&e.relatedTarget instanceof Node&&!l.value.contains(e.relatedTarget)&&k()}function k(){d.value=!1}function x(e){o.value=null;try{const r=ze(e,v.value);r.sort((p,Q)=>p[0].localeCompare(Q[0])),y.value=r}catch(r){if(r instanceof Error)o.value=r,d.value=!0;else throw r}}function R(e,r){return JSON.stringify(e)===JSON.stringify(r)}return(e,r)=>(c(),g("div",{ref_key:"filterBar",ref:l,class:"k-filter-bar","data-testid":"k-filter-bar"},[S("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:K},[qe,f(),S("span",$e,[I(m(ce),{decorative:"","data-testid":"k-filter-bar-filter-icon","hide-title":"",size:m(q)},null,8,["size"])])]),f(),S("label",{for:`${t.id}-filter-bar-input`,class:"visually-hidden"},[se(e.$slots,"default",{},()=>[f(_(u.value),1)],!0)],8,Me),f(),de(S("input",{id:`${t.id}-filter-bar-input`,ref_key:"filterInput",ref:n,"onUpdate:modelValue":r[0]||(r[0]=p=>i.value=p),class:"k-filter-bar-input",type:"text",placeholder:T.value,"data-testid":"k-filter-bar-filter-input",onFocus:r[1]||(r[1]=p=>d.value=!0),onBlur:V,onChange:z},null,40,Ke),[[pe,i.value]]),f(),d.value?(c(),g("div",je,[S("div",Pe,[o.value!==null?(c(),g("p",Ve,_(o.value.message),1)):(c(),g("button",{key:1,class:X(["k-submit-query-button",{"k-submit-query-button-is-selected":b.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:D},`
          Submit `+_(i.value),3)),f(),(c(!0),g(N,null,Y(C.value,(p,Q)=>(c(),g("div",{key:`${t.id}-${Q}`,class:X(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":b.value===Q}])},[S("b",null,_(p.fieldName),1),p.description!==""?(c(),g("span",Re,": "+_(p.description),1)):Z("",!0),f(),S("button",{class:"k-apply-suggestion-button",title:`Add ${p.fieldName}:`,type:"button","data-filter-field":p.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:j},[S("span",Ue,"Add "+_(p.fieldName)+":",1),f(),I(m(fe),{decorative:"","hide-title":"",size:m(q)},null,8,["size"])],8,Qe)],2))),128))])])):Z("",!0),f(),i.value!==""?(c(),g("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:P},[Oe,f(),I(m(me),{decorative:"","hide-title":"",size:m(q)},null,8,["size"])])):Z("",!0)],512))}});const st=ne(He,[["__scopeId","data-v-349996e3"]]);export{tt as D,st as F};
