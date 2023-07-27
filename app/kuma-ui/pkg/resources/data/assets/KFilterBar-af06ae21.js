var ie=Object.defineProperty;var le=(s,l,a)=>l in s?ie(s,l,{enumerable:!0,configurable:!0,writable:!0,value:a}):s[l]=a;var R=(s,l,a)=>(le(s,typeof l!="symbol"?l+"":l,a),a);import{T as re,u as j,a as ue,D as ce}from"./kongponents.es-f12d3d78.js";import{A as de}from"./DataSource.vue_vue_type_script_setup_true_lang-122fb823.js";import{S as pe}from"./StatusBadge-6793ebff.js";import{d as ae,c as M,r as me,o as g,a as V,w as h,x as se,h as w,g as p,t as b,e as k,F as q,q as _,n as G,b as T,K as fe,j as A,C as X,D as ge,E as ve,s as ye,f as Z,k as he,y as be,p as ke,m as _e}from"./index-79776270.js";import{e as Te,g as Se,v as Ce,t as we,r as ee,y as De,C as Ae,B as Ne,E as Ue,z as xe,f as oe}from"./RouteView.vue_vue_type_script_setup_true_lang-bb7b7c66.js";const Ee=ae({__name:"DataPlaneList",props:{total:{default:0},pageNumber:{},pageSize:{},items:{},error:{},gateways:{type:Boolean,default:!1}},emits:["load-data","change"],setup(s,{emit:l}){const a=s,y=Te(),{t:n}=Se(),u=M(()=>y.getters["config/getMulticlusterStatus"]);function d(m){return m.map(r=>{var z,F,x,D,K,t;const c=r.mesh,o=r.name,v=((z=r.dataplane.networking.gateway)==null?void 0:z.type)||"STANDARD",N={name:v==="STANDARD"?"data-plane-detail-view":"gateway-detail-view",params:{mesh:c,dataPlane:o}},O=["kuma.io/protocol","kuma.io/service","kuma.io/zone"],U=Ce(r.dataplane).filter(e=>O.includes(e.label)),E=(F=U.find(e=>e.label==="kuma.io/service"))==null?void 0:F.value,$=(x=U.find(e=>e.label==="kuma.io/protocol"))==null?void 0:x.value,I=(D=U.find(e=>e.label==="kuma.io/zone"))==null?void 0:D.value;let P;E!==void 0&&(P={name:"service-detail-view",params:{mesh:c,service:E}});let L;I!==void 0&&(L={name:"zone-cp-detail-view",params:{zone:I}});const{status:Q}=we(r.dataplane,r.dataplaneInsight),B=((K=r.dataplaneInsight)==null?void 0:K.subscriptions)??[],H={totalUpdates:0,totalRejectedUpdates:0,dpVersion:null,envoyVersion:null,selectedTime:NaN,selectedUpdateTime:NaN,version:null},f=B.reduce((e,i)=>{var J,Y;if(i.connectTime){const W=Date.parse(i.connectTime);(!e.selectedTime||W>e.selectedTime)&&(e.selectedTime=W)}const C=Date.parse(i.status.lastUpdateTime);return C&&(!e.selectedUpdateTime||C>e.selectedUpdateTime)&&(e.selectedUpdateTime=C),{totalUpdates:e.totalUpdates+parseInt(i.status.total.responsesSent??"0",10),totalRejectedUpdates:e.totalRejectedUpdates+parseInt(i.status.total.responsesRejected??"0",10),dpVersion:((J=i.version)==null?void 0:J.kumaDp.version)||e.dpVersion,envoyVersion:((Y=i.version)==null?void 0:Y.envoy.version)||e.envoyVersion,selectedTime:e.selectedTime,selectedUpdateTime:e.selectedUpdateTime,version:i.version||e.version}},H),S={name:o,detailViewRoute:N,type:v,zone:{title:I??"—",route:L},service:{title:E??"—",route:P},protocol:$??"—",status:Q,totalUpdates:f.totalUpdates,totalRejectedUpdates:f.totalRejectedUpdates,dpVersion:f.dpVersion??"—",envoyVersion:f.envoyVersion??"—",warnings:[],unsupportedEnvoyVersion:!1,unsupportedKumaDPVersion:!1,kumaDpAndKumaCpMismatch:!1,lastUpdated:f.selectedUpdateTime?ee(new Date(f.selectedUpdateTime).toUTCString()):"—",lastConnected:f.selectedTime?ee(new Date(f.selectedTime).toUTCString()):"—",overview:r};if(f.version){const{kind:e}=De(f.version);switch(e!==Ae&&S.warnings.push(e),e){case Ue:S.unsupportedEnvoyVersion=!0;break;case Ne:S.unsupportedKumaDPVersion=!0;break}}return u.value&&f.dpVersion&&U.find(i=>i.label===fe)&&typeof((t=f.version)==null?void 0:t.kumaDp.kumaCpCompatible)=="boolean"&&!f.version.kumaDp.kumaCpCompatible&&(S.warnings.push(xe),S.kumaDpAndKumaCpMismatch=!0),S})}return(m,r)=>{const c=me("RouterLink");return g(),V(de,{"empty-state-title":T(n)("common.emptyState.title"),"empty-state-message":T(n)("common.emptyState.message",{type:a.gateways?"Gateways":"Data plane proxies"}),headers:[{label:"Name",key:"name"},a.gateways?{label:"Type",key:"type"}:void 0,{label:"Service",key:"service"},a.gateways?void 0:{label:"Protocol",key:"protocol"},u.value?{label:"Zone",key:"zone"}:void 0,{label:"Last Updated",key:"lastUpdated"},{label:"Kuma DP version",key:"dpVersion"},{label:"Status",key:"status"},{label:"Actions",key:"actions",hideLabel:!0}].filter(Boolean),"page-number":a.pageNumber,"page-size":a.pageSize,total:a.total,items:a.items?d(a.items):void 0,error:a.error,onChange:r[0]||(r[0]=o=>l("change",o))},{toolbar:h(()=>[se(m.$slots,"toolbar",{},void 0,!0)]),name:h(({row:o})=>[w(c,{to:{name:a.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:o.name}},"data-testid":"detail-view-link"},{default:h(()=>[p(b(o.name),1)]),_:2},1032,["to"])]),service:h(({rowValue:o})=>[o.route?(g(),V(c,{key:0,to:o.route},{default:h(()=>[p(b(o.title),1)]),_:2},1032,["to"])):(g(),k(q,{key:1},[p(b(o.title),1)],64))]),dpVersion:h(({row:o,rowValue:v})=>[_("div",{class:G({"with-warnings":o.unsupportedEnvoyVersion||o.unsupportedKumaDPVersion||o.kumaDpAndKumaCpMismatch})},b(v),3)]),zone:h(({rowValue:o})=>[o.route?(g(),V(c,{key:0,to:o.route},{default:h(()=>[p(b(o.title),1)]),_:2},1032,["to"])):(g(),k(q,{key:1},[p(b(o.title),1)],64))]),status:h(({rowValue:o})=>[o?(g(),V(pe,{key:0,status:o},null,8,["status"])):(g(),k(q,{key:1},[p(`
        —
      `)],64))]),actions:h(({row:o})=>[w(T(ce),{class:"actions-dropdown","kpop-attributes":{placement:"bottomEnd",popoverClasses:"mt-5 more-actions-popover"},width:"150"},{default:h(()=>[w(T(re),{class:"non-visual-button",appearance:"secondary",size:"small"},{icon:h(()=>[w(T(j),{color:"var(--black-400)",icon:"more",size:"16"})]),_:1})]),items:h(()=>[w(T(ue),{item:{to:{name:a.gateways?"gateway-detail-view":"data-plane-detail-view",params:{dataPlane:o.name}},label:T(n)("common.collection.actions.view")}},null,8,["item"])]),_:2},1024)]),_:3},8,["empty-state-title","empty-state-message","headers","page-number","page-size","total","items","error"])}}});const st=oe(Ee,[["__scopeId","data-v-c4a80840"]]);function Ie(s,l,a){return Math.max(l,Math.min(s,a))}const Pe=["ControlLeft","ControlRight","ShiftLeft","ShiftRight","AltLeft"];class Me{constructor(l,a){R(this,"commands");R(this,"keyMap");R(this,"boundTriggerShortcuts");this.commands=a,this.keyMap=Object.fromEntries(Object.entries(l).map(([y,n])=>[y.toLowerCase(),n])),this.boundTriggerShortcuts=this.triggerShortcuts.bind(this)}registerListener(){document.addEventListener("keydown",this.boundTriggerShortcuts)}unRegisterListener(){document.removeEventListener("keydown",this.boundTriggerShortcuts)}triggerShortcuts(l){Le(l,this.keyMap,this.commands)}}function Le(s,l,a){const y=Be(s.code),n=[s.ctrlKey?"ctrl":"",s.shiftKey?"shift":"",s.altKey?"alt":"",y].filter(m=>m!=="").join("+"),u=l[n];if(!u)return;const d=a[u];d.isAllowedContext&&!d.isAllowedContext(s)||(d.shouldPreventDefaultAction&&s.preventDefault(),!(d.isDisabled&&d.isDisabled())&&d.trigger(s))}function Be(s){return Pe.includes(s)?"":s.replace(/^Key/,"").toLowerCase()}function ze(s,l){const a=" "+s,y=a.matchAll(/ ([-\s\w]+):\s*/g),n=[];for(const u of Array.from(y)){if(u.index===void 0)continue;const d=Fe(u[1]);if(l.length>0&&!l.includes(d))throw new Error(`Unknown field “${d}”. Known fields: ${l.join(", ")}`);const m=u.index+u[0].length,r=a.substring(m);let c;if(/^\s*["']/.test(r)){const v=r.match(/['"](.*?)['"]/);if(v!==null)c=v[1];else throw new Error(`Quote mismatch for field “${d}”.`)}else{const v=r.indexOf(" "),N=v===-1?r.length:v;c=r.substring(0,N)}c!==""&&n.push([d,c])}return n}function Fe(s){return s.trim().replace(/\s+/g,"-").replace(/-[a-z]/g,(l,a)=>a===0?l:l.substring(1).toUpperCase())}let te=0;const Ke=(s="unique")=>(te++,`${s}-${te}`),ne=s=>(ke("data-v-5a222cc2"),s=s(),_e(),s),Re=ne(()=>_("span",{class:"visually-hidden"},"Focus filter",-1)),Ve=["for"],je=["id","placeholder"],qe={key:0,class:"k-suggestion-box","data-testid":"k-filter-bar-suggestion-box"},Oe={class:"k-suggestion-list"},$e={key:0,class:"k-filter-bar-error"},Qe={key:0},He=["title","data-filter-field"],Ze={class:"visually-hidden"},Ge=ne(()=>_("span",{class:"visually-hidden"},"Clear query",-1)),Je=ae({__name:"KFilterBar",props:{id:{type:String,required:!1,default:()=>Ke("k-filter-bar")},fields:{type:Object,required:!0},placeholder:{type:String,required:!1,default:null},query:{type:String,required:!1,default:""}},emits:["fields-change"],setup(s,{emit:l}){const a=s,y=A(null),n=A(null),u=A(a.query),d=A([]),m=A(null),r=A(!1),c=A(-1),o=M(()=>Object.keys(a.fields)),v=M(()=>Object.entries(a.fields).slice(0,5).map(([t,e])=>({fieldName:t,...e}))),N=M(()=>o.value.length>0?`Filter by ${o.value.join(", ")}`:"Filter"),O=M(()=>a.placeholder??N.value);X(()=>d.value,function(t,e){K(t,e)||(m.value=null,l("fields-change",{fields:t,query:u.value}))}),X(()=>u.value,function(){u.value===""&&(m.value=null),r.value=!0});const U={Enter:"submitQuery",Escape:"closeSuggestionBox",ArrowDown:"jumpToNextSuggestion",ArrowUp:"jumpToPreviousSuggestion"},E={submitQuery:{trigger:P,isAllowedContext(t){return n.value!==null&&t.composedPath().includes(n.value)},shouldPreventDefaultAction:!0},jumpToNextSuggestion:{trigger:L,isAllowedContext(t){return n.value!==null&&t.composedPath().includes(n.value)},shouldPreventDefaultAction:!0},jumpToPreviousSuggestion:{trigger:Q,isAllowedContext(t){return n.value!==null&&t.composedPath().includes(n.value)},shouldPreventDefaultAction:!0},closeSuggestionBox:{trigger:x,isAllowedContext(t){return y.value!==null&&t.composedPath().includes(y.value)}}};function $(){const t=new Me(U,E);he(function(){t.registerListener()}),be(function(){t.unRegisterListener()}),D(u.value)}$();function I(t){const e=t.target;D(e.value)}function P(){if(n.value instanceof HTMLInputElement)if(c.value===-1)D(n.value.value),r.value=!1;else{const t=v.value[c.value].fieldName;t&&S(n.value,t)}}function L(){B(1)}function Q(){B(-1)}function B(t){c.value=Ie(c.value+t,-1,v.value.length-1)}function H(){n.value instanceof HTMLInputElement&&n.value.focus()}function f(t){const i=t.currentTarget.getAttribute("data-filter-field");i&&n.value instanceof HTMLInputElement&&S(n.value,i)}function S(t,e){const i=u.value===""||u.value.endsWith(" ")?"":" ";u.value+=i+e+":",t.focus(),c.value=-1}function z(){u.value="",n.value instanceof HTMLInputElement&&(n.value.value="",n.value.focus(),D(""))}function F(t){t.relatedTarget===null&&x(),y.value instanceof HTMLElement&&t.relatedTarget instanceof Node&&!y.value.contains(t.relatedTarget)&&x()}function x(){r.value=!1}function D(t){m.value=null;try{const e=ze(t,o.value);e.sort((i,C)=>i[0].localeCompare(C[0])),d.value=e}catch(e){if(e instanceof Error)m.value=e,r.value=!0;else throw e}}function K(t,e){return JSON.stringify(t)===JSON.stringify(e)}return(t,e)=>(g(),k("div",{ref_key:"filterBar",ref:y,class:"k-filter-bar","data-testid":"k-filter-bar"},[_("button",{class:"k-focus-filter-input-button",title:"Focus filter",type:"button","data-testid":"k-filter-bar-focus-filter-input-button",onClick:H},[Re,p(),w(T(j),{"aria-hidden":"true",class:"k-filter-icon",color:"var(--grey-400)","data-testid":"k-filter-bar-filter-icon","hide-title":"",icon:"filter",size:"20"})]),p(),_("label",{for:`${a.id}-filter-bar-input`,class:"visually-hidden"},[se(t.$slots,"default",{},()=>[p(b(N.value),1)],!0)],8,Ve),p(),ge(_("input",{id:`${a.id}-filter-bar-input`,ref_key:"filterInput",ref:n,"onUpdate:modelValue":e[0]||(e[0]=i=>u.value=i),class:"k-filter-bar-input",type:"text",placeholder:O.value,"data-testid":"k-filter-bar-filter-input",onFocus:e[1]||(e[1]=i=>r.value=!0),onBlur:F,onChange:I},null,40,je),[[ve,u.value]]),p(),r.value?(g(),k("div",qe,[_("div",Oe,[m.value!==null?(g(),k("p",$e,b(m.value.message),1)):(g(),k("button",{key:1,class:G(["k-submit-query-button",{"k-submit-query-button-is-selected":c.value===-1}]),title:"Submit query",type:"button","data-testid":"k-filter-bar-submit-query-button",onClick:P},`
          Submit `+b(u.value),3)),p(),(g(!0),k(q,null,ye(v.value,(i,C)=>(g(),k("div",{key:`${a.id}-${C}`,class:G(["k-suggestion-list-item",{"k-suggestion-list-item-is-selected":c.value===C}])},[_("b",null,b(i.fieldName),1),i.description!==""?(g(),k("span",Qe,": "+b(i.description),1)):Z("",!0),p(),_("button",{class:"k-apply-suggestion-button",title:`Add ${i.fieldName}:`,type:"button","data-filter-field":i.fieldName,"data-testid":"k-filter-bar-apply-suggestion-button",onClick:f},[_("span",Ze,"Add "+b(i.fieldName)+":",1),p(),w(T(j),{"aria-hidden":"true",color:"currentColor","hide-title":"",icon:"chevronRight",size:"16"})],8,He)],2))),128))])])):Z("",!0),p(),u.value!==""?(g(),k("button",{key:1,class:"k-clear-query-button",title:"Clear query",type:"button","data-testid":"k-filter-bar-clear-query-button",onClick:z},[Ge,p(),w(T(j),{"aria-hidden":"true",color:"currentColor",icon:"clear","hide-title":"",size:"20"})])):Z("",!0)],512))}});const ot=oe(Je,[["__scopeId","data-v-5a222cc2"]]);export{st as D,ot as K};
