var ie=Object.defineProperty;var le=(s,i,o)=>i in s?ie(s,i,{enumerable:!0,configurable:!0,writable:!0,value:o}):s[i]=o;var K=(s,i,o)=>(le(s,typeof i!="symbol"?i+"":i,o),o);import{d as oe,c as M,r as re,o as p,a as z,w as y,z as ae,h as w,g as d,t as _,e as k,F as E,b as v,L as ue,v as P,V as ce,D as de,H as pe,K as me,j as D,J as X,q as T,O as fe,P as ge,n as ee,s as ve,f as J,k as ye,A as he,p as be,m as ke}from"./index-0105b00b.js";import{A as _e}from"./AppCollection-930d9b16.js";import{e as Se,g as Te,S as we,f as ne}from"./RouteView.vue_vue_type_script_setup_true_lang-fff04e01.js";import{d as Ce,a as Ae,c as De,C as xe,e as Ne}from"./dataplane-30467516.js";import{n as Ue}from"./notEmpty-7f452b20.js";const Ie=oe({__name:"DataPlaneList",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{},gateways:{type:Boolean,default:!1}},emits:["load-data","change"],setup(s,{emit:i}){const o=s,h=Se(),{t:a,formatIsoDate:u}=Te(),c=M(()=>h.getters["config/getMulticlusterStatus"]);function b(m){return m.map(r=>{var R,U,A,q,t,l;const S=r.mesh,n=r.name,C=((R=r.dataplane.networking.gateway)==null?void 0:R.type)||"STANDARD",$={name:C==="STANDARD"?"data-plane-detail-view":"gateway-detail-view",params:{mesh:S,dataPlane:n}},V=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],x=Ce(r.dataplane).filter(e=>V.includes(e.label)),I=(U=x.find(e=>e.label==="kuma.io/service"))==null?void 0:U.value,O=(A=x.find(e=>e.label==="kuma.io/protocol"))==null?void 0:A.value,N=(q=x.find(e=>e.label==="kuma.io/zone"))==null?void 0:q.value;let F;I!==void 0&&(F={name:"service-detail-view",params:{mesh:S,service:I}});let j;N!==void 0&&(j={name:"zone-cp-detail-view",params:{zone:N}});const{status:B}=Ae(r.dataplane,r.dataplaneInsight),Q=((t=r.dataplaneInsight)==null?void 0:t.subscriptions)??[],H={totalUpdates:0,totalRejectedUpdates:0,dpVersion:null,envoyVersion:null,selectedTime:NaN,selectedUpdateTime:NaN,version:null},f=Q.reduce((e,g)=>{var G,W;if(g.connectTime){const Y=Date.parse(g.connectTime);(!e.selectedTime||Y>e.selectedTime)&&(e.selectedTime=Y)}const Z=Date.parse(g.status.lastUpdateTime);return Z&&(!e.selectedUpdateTime||Z>e.selectedUpdateTime)&&(e.selectedUpdateTime=Z),{totalUpdates:e.totalUpdates+parseInt(g.status.total.responsesSent??"0",10),totalRejectedUpdates:e.totalRejectedUpdates+parseInt(g.status.total.responsesRejected??"0",10),dpVersion:((G=g.version)==null?void 0:G.kumaDp.version)||e.dpVersion,envoyVersion:((W=g.version)==null?void 0:W.envoy.version)||e.envoyVersion,selectedTime:e.selectedTime,selectedUpdateTime:e.selectedUpdateTime,version:g.version||e.version}},H),L={name:n,detailViewRoute:$,type:C,zone:{title:N??a("common.collection.none"),route:j},service:{title:I??a("common.collection.none"),route:F},protocol:O??a("common.collection.none"),status:B,totalUpdates:f.totalUpdates,totalRejectedUpdates:f.totalRejectedUpdates,envoyVersion:f.envoyVersion??a("common.collection.none"),warnings:[],lastUpdated:f.selectedUpdateTime?u(new Date(f.selectedUpdateTime).toUTCString()):a("common.collection.none"),lastConnected:f.selectedTime?u(new Date(f.selectedTime).toUTCString()):a("common.collection.none"),overview:r};if(f.version){const{kind:e}=De(f.version);e!==xe&&L.warnings.push(e)}return c.value&&f.dpVersion&&x.find(g=>g.label===me)&&typeof((l=f.version)==null?void 0:l.kumaDp.kumaCpCompatible)=="boolean"&&!f.version.kumaDp.kumaCpCompatible&&L.warnings.push(Ne),L})}return(m,r)=>{const S=re("RouterLink");return p(),z(_e,{"empty-state-title":v(a)("common.emptyState.title"),"empty-state-message":v(a)("common.emptyState.message",{type:o.gateways?"Gateways":"Data Plane Proxies"}),headers:[{label:"Name",key:"name"},o.gateways?{label:"Type",key:"type"}:void 0,{label:"Service",key:"service"},o.gateways?void 0:{label:"Protocol",key:"protocol"},c.value?{label:"Zone",key:"zone"}:void 0,{label:"Last Updated",key:"lastUpdated"},{label:"Status",key:"status"},{label:"Warnings",key:"warnings",hideLabel:!0},{label:"Actions",key:"actions",hideLabel:!0}].filter(v(Ue)),"page-number":o.pageNumber,"page-size":o.pageSize,total:o.total,items:o.items?b(o.items):void 0,error:o.error,onChange:r[0]||(r[0]=n=>i("change",n))},{toolbar:y(()=>[ae(m.$slots,"toolbar",{},void 0,!0)]),name:y(({row:n})=>[w(S,{to:{name:o.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:n.name}},"data-testid":"detail-view-link"},{default:y(()=>[d(_(n.name),1)]),_:2},1032,["to"])]),service:y(({rowValue:n})=>[n.route?(p(),z(S,{key:0,to:n.route},{default:y(()=>[d(_(n.title),1)]),_:2},1032,["to"])):(p(),k(E,{key:1},[d(_(n.title),1)],64))]),zone:y(({rowValue:n})=>[n.route?(p(),z(S,{key:0,to:n.route},{default:y(()=>[d(_(n.title),1)]),_:2},1032,["to"])):(p(),k(E,{key:1},[d(_(n.title),1)],64))]),status:y(({rowValue:n})=>[n?(p(),z(we,{key:0,status:n},null,8,["status"])):(p(),k(E,{key:1},[d(_(v(a)("common.collection.none")),1)],64))]),warnings:y(({rowValue:n})=>[n.length>0?(p(),z(v(ue),{key:0,label:v(a)("data-planes.list.version_mismatch")},{default:y(()=>[w(v(P),{class:"mr-1",icon:"warning",color:"var(--black-500)","secondary-color":"var(--yellow-300)",size:"20","hide-title":""})]),_:1},8,["label"])):(p(),k(E,{key:1},[d(`
         
      `)],64))]),actions:y(({row:n})=>[w(v(ce),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:y(()=>[w(v(de),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:y(()=>[w(v(P),{color:"var(--black-400)",icon:"more",size:"16"})]),_:1})]),items:y(()=>[w(v(pe),{item:{to:{name:o.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:n.name}},label:v(a)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:3},8,["empty-state-title","empty-state-message","headers","page-number","page-size","total","items","error"])}}});const at=ne(Ie,[["__scopeId","data-v-66c9dcb2"]]);function Le(s,i,o){return Math.max(i,Math.min(s,o))}const ze=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class Me{constructor(i,o){K(this,"commands");K(this,"keyMap");K(this,"boundTriggerShortcuts");this.commands=o,this.keyMap=Object.fromEntries(Object.entries(i).map(([h,a])=>[h.toLowerCase(),a])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(i){Ee(i,this.keyMap,this.commands)}}function Ee(s,i,o){const h=Pe(s.code),a=[s.ctrlKey?"ctrl":"",s.shiftKey?"shift":"",s.altKey?"alt":"",h].filter(b=>b!=="").join("+"),u=i[a];if(!u)return;const c=o[u];c.isAllowedContext&&!c.isAllowedContext(s)||(c.shouldPreventDefaultAction&&s.preventDefault(),!(c.isDisabled&&c.isDisabled())&&c.trigger(s))}function Pe(s){return ze.includes(s)?"":s.replace(/^Key/,"").toLowerCase()}function Fe(s,i){const o=" "+s,h=o.matchAll(/ ([-\s\w]+):\s*/g),a=[];for(const u of Array.from(h)){if(u.index===void 0)continue;const c=je(u[1]);if(i.length>0&&!i.includes(c))throw new Error(`Unknown field “${c}”. Known fields: ${i.join(", ")}`);const b=u.index+u[0].length,m=o.substring(b);let r;if(/^\s*["']/.test(m)){const n=m.match(/['"](.*?)['"]/);if(n!==null)r=n[1];else throw new Error(`Quote mismatch for field “${c}”.`)}else{const n=m.indexOf(" "),C=n===-1?m.length:n;r=m.substring(0,C)}r!==""&&a.push([c,r])}return a}function je(s){return s.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(i,o)=>o===0?i:i.substring(1).toUpperCase())}let te=0;const Be=(s="unique")=>(te++,`${s}-${te}`),se=s=>(be("data-v-121f7a4c"),s=s(),ke(),s),Re=se(()=>T("span",{class:"visually-hidden"},"Focus filter",-1)),qe=["for"],Ke=["id","placeholder"],$e={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},Ve={class:"k-suggestion-list"},Oe={key:0,class:"k-filter-bar-error"},Qe={key:0},He=["title","data-filter-field"],Ze={class:"visually-hidden"},Je=se(()=>T("span",{class:"visually-hidden"},"Clear query",-1)),Ge=oe({__name:"KFilterBar",props:{id:{type:String,required:!1,default:()=>Be("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(s,{emit:i}){const o=s,h=D(null),a=D(null),u=D(o.query),c=D([]),b=D(null),m=D(!1),r=D(-1),S=M(()=>Object.keys(o.fields)),n=M(()=>Object.entries(o.fields).slice(0,5).map(([t,l])=>({fieldName:t,...l}))),C=M(()=>S.value.length>0?`Filter by ${S.value.join(", ")}`:"Filter"),$=M(()=>o.placeholder??C.value);X(()=>c.value,function(t,l){q(t,l)||(b.value=null,i("fields-change",{fields:t,query:u.value}))}),X(()=>u.value,function(){u.value===""&&(b.value=null),m.value=!0});const V={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},x={submitQuery:{trigger:N,isAllowedContext(t){return a.value!==null&&t.composedPath().includes(a.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:F,isAllowedContext(t){return a.value!==null&&t.composedPath().includes(a.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:j,isAllowedContext(t){return a.value!==null&&t.composedPath().includes(a.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:U,isAllowedContext(t){return h.value!==null&&t.composedPath().includes(h.value)}}};function I(){const t=new Me(V,x);ye(function(){t.registerListener()}),he(function(){t.unRegisterListener()}),A(u.value)}I();function O(t){const l=t.target;A(l.value)}function N(){if(a.value instanceof HTMLInputElement)if(r.value===-1)A(a.value.value),m.value=!1;else{const t=n.value[r.value].fieldName;t&&f(a.value,t)}}function F(){B(1)}function j(){B(-1)}function B(t){r.value=Le(r.value+t,-1,n.value.length-1)}function Q(){a.value instanceof HTMLInputElement&&a.value.focus()}function H(t){const e=t.currentTarget.getAttribute("data-filter-field");e&&a.value instanceof HTMLInputElement&&f(a.value,e)}function f(t,l){const e=u.value===""||u.value.endsWith(" ")?"":" ";u.value+=e+l+":",t.focus(),r.value=-1}function L(){u.value="",a.value instanceof HTMLInputElement&&(a.value.value="",a.value.focus(),A(""))}function R(t){t.relatedTarget===null&&U(),h.value instanceof HTMLElement&&t.relatedTarget instanceof Node&&!h.value.contains(t.relatedTarget)&&U()}function U(){m.value=!1}function A(t){b.value=null;try{const l=Fe(t,S.value);l.sort((e,g)=>e[0].localeCompare(g[0])),c.value=l}catch(l){if(l instanceof Error)b.value=l,m.value=!0;else throw l}}function q(t,l){return JSON.stringify(t)===JSON.stringify(l)}return(t,l)=>(p(),k("div",{ref_key:"filterBar",ref:h,class:"k-filter-bar","data-testid":"k-filter-bar"},[T("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:Q},[Re,d(),w(v(P),{"aria-hidden":"true",class:"k-filter-icon",color:"var(--grey-400)","data-testid":"k-filter-bar-filter-icon","hide-title":"",icon:"filter",size:"20"})]),d(),T("label",{for:`${o.id}-filter-bar-input`,class:"visually-hidden"},[ae(t.$slots,"default",{},()=>[d(_(C.value),1)],!0)],8,qe),d(),fe(T("input",{id:`${o.id}-filter-bar-input`,ref_key:"filterInput",ref:a,"onUpdate:modelValue":l[0]||(l[0]=e=>u.value=e),class:"k-filter-bar-input",type:"text",placeholder:$.value,"data-testid":"k-filter-bar-filter-input",onFocus:l[1]||(l[1]=e=>m.value=!0),onBlur:R,onChange:O},null,40,Ke),[[ge,u.value]]),d(),m.value?(p(),k("div",$e,[T("div",Ve,[b.value!==null?(p(),k("p",Oe,_(b.value.message),1)):(p(),k("button",{key:1,class:ee(["k-submit-query-button",{"k-submit-query-button-is-selected":r.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:N},`
          Submit `+_(u.value),3)),d(),(p(!0),k(E,null,ve(n.value,(e,g)=>(p(),k("div",{key:`${o.id}-${g}`,class:ee(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":r.value===g}])},[T("b",null,_(e.fieldName),1),e.description!==""?(p(),k("span",Qe,": "+_(e.description),1)):J("",!0),d(),T("button",{class:"k-apply-suggestion-button",title:`Add ${e.fieldName}:`,type:"button","data-filter-field":e.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:H},[T("span",Ze,"Add "+_(e.fieldName)+":",1),d(),w(v(P),{"aria-hidden":"true",color:"currentColor","hide-title":"",icon:"chevronRight",size:"16"})],8,He)],2))),128))])])):J("",!0),d(),u.value!==""?(p(),k("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:L},[Je,d(),w(v(P),{"aria-hidden":"true",color:"currentColor",icon:"clear","hide-title":"",size:"20"})])):J("",!0)],512))}});const nt=ne(Ge,[["__scopeId","data-v-121f7a4c"]]);export{at as D,nt as K};
